package http_test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	fluxhttp "github.com/influxdata/flux/dependencies/http"
	"github.com/influxdata/flux/dependencies/url"
	_ "github.com/influxdata/flux/fluxinit/static"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/runtime"
)

func TestPost(t *testing.T) {
	var req *http.Request
	var body []byte
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req = r
		var err error
		body, err = ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
		}
		w.WriteHeader(204)
	}))
	defer ts.Close()

	script := fmt.Sprintf(`
import "http"
import "internal/testutil"

status = http.post(url:"%s/path/a/b/c", headers: {x:"a",y:"b",z:"c"}, data: bytes(v: "body"))
status == 204 or testutil.fail()
`, ts.URL)

	ctx := flux.NewDefaultDependencies().Inject(context.Background())
	if _, _, err := runtime.Eval(ctx, script); err != nil {
		t.Fatal("evaluation of http.post failed: ", err)
	}
	if want, got := "/path/a/b/c", req.URL.Path; want != got {
		t.Errorf("unexpected url want: %q got: %q", want, got)
	}
	if want, got := "POST", req.Method; want != got {
		t.Errorf("unexpected method want: %q got: %q", want, got)
	}
	header := make(http.Header)
	header.Set("x", "a")
	header.Set("y", "b")
	header.Set("z", "c")
	header.Set("Accept-Encoding", "gzip")
	header.Set("Content-Length", "4")
	header.Set("User-Agent", "Go-http-client/1.1")
	if !cmp.Equal(header, req.Header) {
		t.Errorf("unexpected header -want/+got\n%s", cmp.Diff(header, req.Header))
	}

	expBody := []byte("body")
	if !bytes.Equal(body, expBody) {
		t.Errorf("unexpected body want: %q got: %q", string(expBody), string(body))
	}
}

func TestPost_ValidationFail(t *testing.T) {
	script := `
import "http"

http.post(url:"http://127.1.1.1/path/a/b/c", headers: {x:"a",y:"b",z:"c"}, data: bytes(v: "body"))
`

	deps := flux.NewDefaultDependencies()
	deps.Deps.HTTPClient = fluxhttp.NewDefaultClient(url.PrivateIPValidator{})
	ctx := deps.Inject(context.Background())
	_, _, err := runtime.Eval(ctx, script)
	if err == nil {
		t.Fatal("expected failure")
	}
	if !strings.Contains(err.Error(), "url is not valid") {
		t.Errorf("unexpected cause of failure, got err: %v", err)
	}
	// At time of writing, the initial Error should be a 0 (inherit), with an inner 3 (invalid)
	if code := err.(*errors.Error).Err.(*errors.Error).Code; code != codes.Invalid {
		t.Errorf("unexpected error code. Wanted %q got %q", codes.Invalid, code)
	}
}
