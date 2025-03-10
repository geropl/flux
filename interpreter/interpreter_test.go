package interpreter_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/repl"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/semantic/semantictest"
	"github.com/influxdata/flux/values"
)

var prelude = `
import "internal/testutil"
import "dict"
get = dict.get
insert = dict.insert
fail = testutil.fail
fortyTwo = () => 42.0
six = () => 6.0
nine = () => 9.0
plusOne = (x=<-) => x + 1.0
makeRecord = testutil.makeRecord
hasValue = (o) => exists o.value
sideEffect = () => 0 |> testutil.yield()
`

// TestEval tests whether a program can run to completion or not
func TestEval(t *testing.T) {
	any := `.+`
	testCases := []struct {
		name    string
		query   string
		wantErr string
		want    []values.Value
	}{
		{
			name: "string interpolation",
			query: `
				str = "str"
				ing = "ing"
				"str + ing = ${str+ing}"`,
			want: []values.Value{
				values.NewString("str + ing = string"),
			},
		},
		{
			name: "string interpolation missing field",
			query: `
				r = makeRecord(o: {a: "foo", b: 42})
				"r._value = ${r._value}"`,
			wantErr: any,
		},
		{
			name: "string interpolation non-string type",
			query: `
				r = makeRecord(o: {a: "foo", b: 42})
				"r._value = ${r.b}"`,
			want: []values.Value{
				values.NewString("r._value = 42"),
			},
		},
		{
			name: "string interpolation wrong type",
			query: `
				r = makeRecord(o: {a: "foo", b: 42})
				"r = ${r}"`,
			wantErr: any,
		},
		{
			name:  "call builtin function",
			query: "six()",
			want: []values.Value{
				values.NewFloat(6.0),
			},
		},
		{
			name:    "call function with fail",
			query:   "fail()",
			wantErr: any,
		},
		{
			name:    "call function with duplicate args",
			query:   "plusOne(x:1.0, x:2.0)",
			wantErr: any,
		},
		{
			name: "binary expressions",
			query: `
			six_value = six()
			nine_value = nine()

			fortyTwo() == six_value * nine_value
			`,
			want: []values.Value{
				values.NewBool(false),
			},
		},
		{
			name: "logical expressions short circuit",
			query: `
            six_value = six()
            nine_value = nine()

            not (fortyTwo() == six_value * nine_value) or fail()
			`,
			want: []values.Value{
				values.NewBool(true),
			},
		},
		{
			name: "function",
			query: `
            plusSix = (r) => r + six()
            plusSix(r:1.0) == 7.0 or fail()
			`,
		},
		{
			name: "function block",
			query: `
            f = (r) => {
                r1 = 1.0 + r
                return (r + r1) / r
            }
            f(r:1.0) == 3.0 or fail()
			`,
		},
		{
			name: "function block polymorphic",
			query: `
            f = (r) => {
                r2 = r * r
                return r2 / r
            }
            f(r:2.0) == 2.0 or fail()
            f(r:2) == 2 or fail()
			`,
		},
		{
			name: "function with default param",
			query: `
            addN = (r,n=4) => r + n
            addN(r:2) == 6 or fail()
            addN(r:3,n:1) == 4 or fail()
			`,
		},
		{
			name: "scope closing",
			query: `
			x = 5
            plusX = (r) => r + x
            plusX(r:2) == 7 or fail()
			`,
		},
		{
			name: "nested scope mutations not visible outside",
			query: `
			x = 5
            xinc = () => {
                x = x + 1
                return x
            }
            xinc() == 6 or fail()
            x == 5 or fail()
			`,
		},
		// TODO(jsternberg): This test seems to not
		// infer the type constraints correctly for m.a,
		// but it doesn't fail.
		{
			name: "return map from func",
			query: `
            toMap = (a,b) => ({
                a: a,
                b: b,
            })
            m = toMap(a:1, b:false)
            m.a == 1 or fail()
            not m.b or fail()
			`,
		},
		{
			name: "pipe expression",
			query: `
			add = (a=<-,b) => a + b
			one = 1
			one |> add(b:2) == 3 or fail()
			`,
		},
		{
			name: "ignore pipe default",
			query: `
			add = (a=<-,b) => a + b
			add(a:1, b:2) == 3 or fail()
			`,
		},
		{
			name: "pipe expression function",
			query: `
			add = (a=<-,b) => a + b
			six() |> add(b:2.0) == 8.0 or fail()
			`,
		},
		{
			name: "pipe builtin function",
			query: `
			six() |> plusOne() == 7.0 or fail()
			`,
			want: []values.Value{
				values.NewBool(true),
			},
		},
		{
			name: "regex match",
			query: `
			"abba" =~ /^a.*a$/ or fail()
			`,
			want: []values.Value{
				values.NewBool(true),
			},
		},
		{
			name: "regex not match",
			query: `
			"abc" =~ /^a.*a$/ and fail()
			`,
			want: []values.Value{
				values.NewBool(false),
			},
		},
		{
			name: "not regex match",
			query: `
			"abc" !~ /^a.*a$/ or fail()
			`,
			want: []values.Value{
				values.NewBool(true),
			},
		},
		{
			name: "not regex not match",
			query: `
			"abba" !~ /^a.*a$/ and fail()
			`,
			want: []values.Value{
				values.NewBool(false),
			},
		},
		{
			name: "options metadata",
			query: `
			option task = {
				name: "foo",
				repeat: 100,
			}
			task.name == "foo" or fail()
			task.repeat == 100 or fail()
			`,
			want: []values.Value{
				values.NewBool(true),
				values.NewBool(true),
			},
		},
		{
			name:  "query with side effects",
			query: `sideEffect() == 0 or fail()`,
			want: []values.Value{
				values.NewInt(0),
				values.NewBool(true),
			},
		},
		{
			name: "array index expression",
			query: `
				a = [1, 2, 3]
				x = a[1]
				x == 2 or fail()
			`,
		},
		{
			name: "dict expression",
			query: `
				m = ["a" + "b": 0, "c" + "d": 1]
				x = get(dict: m, key: "ab", default: 2)
				y = get(dict: m, key: "cd", default: 2)
				z = get(dict: m, key: "ef", default: 2)
				x == 0 and y == 1 and z == 2 or fail()
			`,
		},
		{
			name: "empy dictionary",
			query: `
				m0 = [:]
				m1 = insert(dict: m0, key: "a", value: 0)
				m2 = insert(dict: m0, key: 0, value: "a")
				v1 = get(dict: m1, key: "a", default: -1)
				v2 = get(dict: m2, key: 0, default: "b")
				v1 == 0 and v2 == "a" or fail()
			`,
		},
		{
			name: "array index expression out of bounds low",
			query: `
				a = [1, 2, 3]
				i = -1
				x = a[i]
			`,
			wantErr: any,
		},
		{
			name: "array index expression out of bounds high",
			query: `
				a = [1, 2, 3]
				i = 3
				x = a[i]
			`,
			wantErr: any,
		},
		{
			name: "array with complex index expression",
			query: `
				f = () => ({l: 0, m: 1, n: 2})
				a = [1, 2, 3]
				x = a[f().l]
				y = a[f().m]
				z = a[f().n]
				x == 1 or fail()
				y == 2 or fail()
				z == 3 or fail()
			`,
		},
		{
			name: "short circuit logical and",
			query: `
                false and fail()
            `,
			want: []values.Value{
				values.NewBool(false),
			},
		},
		{
			name: "short circuit logical or",
			query: `
                true or fail()
            `,
			want: []values.Value{
				values.NewBool(true),
			},
		},
		{
			name: "no short circuit logical and",
			query: `
                true and fail()
            `,
			wantErr: any,
		},
		{
			name: "no short circuit logical or",
			query: `
                false or fail()
            `,
			wantErr: any,
		},
		{
			name: "conditional true",
			query: `
				if 1 != 0 then 10 else 100
			`,
			want: []values.Value{
				values.NewInt(10),
			},
		},
		{
			name: "conditional false",
			query: `
				if 1 == 0 then 10 else 100
			`,
			want: []values.Value{
				values.NewInt(100),
			},
		},
		{
			name: "conditional in function",
			query: `
				f = (t, c, a) => if t then c else a
				{
					v1: f(t: false, c: 30, a: 300),
					v2: f(t: true, c: "cats", a: "dogs"),
				}
			`,
			want: []values.Value{
				values.NewObjectWithValues(map[string]values.Value{
					"v1": values.NewInt(300),
					"v2": values.NewString("cats"),
				}),
			},
		},
		{
			name:  "exists",
			query: `hasValue(o: makeRecord(o: {value: 1}))`,
			want: []values.Value{
				values.NewBool(true),
			},
		},
		{
			name:  "exists null",
			query: `hasValue(o: makeRecord(o: {val: 2}))`,
			want: []values.Value{
				values.NewBool(false),
			},
		},
		{
			name:    "invalid function parameter",
			query:   `from(bucket: "telegraf") |> window(every: 0s)`,
			wantErr: `error calling function "window" @\d+:\d+-\d+:\d+: window function requires at least one of "every" or "period" to be set and non-zero`,
		},
		{
			// tests that we don't nest error messages when
			// a function call fails and gets piped into another
			// function.
			name:    "nested function error",
			query:   `from(bucket: "telegraf") |> window(every: 0s) |> mean()`,
			wantErr: `error calling function "window" @\d+:\d+-\d+:\d+: window function requires at least one of "every" or "period" to be set and non-zero`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			src := prelude + tc.query

			ctx := dependenciestest.Default().Inject(context.Background())
			sideEffects, _, err := runtime.Eval(ctx, src)
			if err != nil {
				if tc.wantErr == "" {
					t.Fatalf("unexpected error: %s", err)
				}

				// We expect an error, so it should be a non-internal Flux error.
				if code := flux.ErrorCode(err); code == codes.Internal || code == codes.Unknown {
					t.Errorf("expected non-internal error code, got %s", code)
				}

				re := regexp.MustCompile(tc.wantErr)
				if got := err.Error(); !re.MatchString(got) {
					t.Errorf("expected error to match pattern %q, but error was %q", tc.wantErr, got)
				}
				return
			} else if tc.wantErr != "" {
				t.Fatal("expected error")
			}

			vs := getSideEffectsValues(sideEffects)
			if tc.want != nil && !cmp.Equal(tc.want, vs, semantictest.CmpOptions...) {
				t.Fatalf("unexpected side effect values -want/+got: \n%s", cmp.Diff(tc.want, vs, semantictest.CmpOptions...))
			}
		})
	}
}
func TestEval_Operator_Precedence(t *testing.T) {
	testCases := []struct {
		src  string
		want values.Value
	}{
		{
			src:  "2.0 * 3.0 ^ 2.0",
			want: values.NewFloat(18.0),
		},
		{
			src:  "(2.0 * 3.0) ^ 2.0",
			want: values.NewFloat(36.0),
		},
		{
			src:  "4.0 / 2.0 ^ 2.0",
			want: values.NewFloat(1.0),
		},
		{
			src:  "(4.0 / 2.0) ^ 2.0",
			want: values.NewFloat(4.0),
		},
		{
			src:  "2.0 % 4.0 ^ 2.0",
			want: values.NewFloat(2.0),
		},
		{
			src:  "(2.0 % 4.0) ^ 2.0",
			want: values.NewFloat(4.0),
		},
		{
			src:  "1.0 + 2.0 * 3.0",
			want: values.NewFloat(7.0),
		},
		{
			src:  "(1.0 + 2.0) * 3.0",
			want: values.NewFloat(9.0),
		},
		{
			src:  "1.0 - 2.0 * 3.0",
			want: values.NewFloat(-5.0),
		},
		{
			src:  "(1.0 - 2.0) * 3.0",
			want: values.NewFloat(-3.0),
		},
		{
			src:  "1.0 + 4.0 / 2.0",
			want: values.NewFloat(3.0),
		},
		{
			src:  "(1.0 + 4.0) / 2.0",
			want: values.NewFloat(2.5),
		},
		{
			src:  "1.0 - 4.0 / 2.0",
			want: values.NewFloat(-1.0),
		},
		{
			src:  "(1.0 - 4.0) / 2.0",
			want: values.NewFloat(-1.5),
		},
		{
			src:  "1.0 + 2.0 + 3.0",
			want: values.NewFloat(6.0),
		},
		{
			src:  "1.0 + (2.0 + 3.0)",
			want: values.NewFloat(6.0),
		},
		{
			src:  "1.0 + 2.0 < 4.0",
			want: values.NewBool(true),
		},
		{
			src:  "(1.0 + 2.0) < 4.0",
			want: values.NewBool(true),
		},
		{
			src:  "1.0 + 2.0 <= 4.0",
			want: values.NewBool(true),
		},
		{
			src:  "(1.0 + 2.0) <= 4.0",
			want: values.NewBool(true),
		},
		{
			src:  "1.0 + 2.0 > 4.0",
			want: values.NewBool(false),
		},
		{
			src:  "(1.0 + 2.0) > 4.0",
			want: values.NewBool(false),
		},
		{
			src:  "1.0 + 2.0 >= 4.0",
			want: values.NewBool(false),
		},
		{
			src:  "(1.0 + 2.0) >= 4.0",
			want: values.NewBool(false),
		},
		{
			src:  "false and true or true",
			want: values.NewBool(true),
		},
		{
			src:  "false and (true or true)",
			want: values.NewBool(false),
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.src, func(t *testing.T) {
			ctx := dependenciestest.Default().Inject(context.Background())
			sideEffects, _, err := runtime.Eval(ctx, tc.src)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			vs := getSideEffectsValues(sideEffects)
			want := []values.Value{tc.want}
			if !cmp.Equal(want, vs, semantictest.CmpOptions...) {
				t.Fatalf("unexpected side effect values -want/+got: \n%s", cmp.Diff(want, vs, semantictest.CmpOptions...))
			}
		})
	}
}

func TestInterpreter_MultiPhaseInterpretation(t *testing.T) {
	testCases := []struct {
		name     string
		builtins []string
		program  string
		wantErr  bool
		want     []values.Value
	}{
		{
			// Evaluate two builtin functions in a single phase
			name: "2-phase interpretation",
			builtins: []string{
				`
					_highestOrLowest = (table=<-, reducer) => table |> reducer()
					highestCurrent = (table=<-) => table |> _highestOrLowest(reducer: (table=<-) => table)
				`,
			},
			program: `5 |> highestCurrent()`,
		},
		{
			// Evaluate two builtin functions each in a separate phase
			name: "3-phase interpretation",
			builtins: []string{
				`_highestOrLowest = (table=<-, reducer) => table |> reducer()`,
				`highestCurrent = (table=<-) => table |> _highestOrLowest(reducer: (table=<-) => table)`,
			},
			program: `5 |> highestCurrent()`,
		},
		{
			// Type-check function expression even though it is not called
			// Program is correctly typed so it should not throw any type errors
			name:     "builtin not called - no type error",
			builtins: []string{`_highestOrLowest = (table=<-, reducer) => table |> reducer()`},
			program:  `f = () => 5 |> _highestOrLowest(reducer: (table=<-) => table)`,
		},
		{
			// Type-check function expression even though it is not called
			// Program should not type check due to missing pipe parameter
			name:     "builtin not called - type error",
			builtins: []string{`_highestOrLowest = (table=<-) => table`},
			program:  `f = () => _highestOrLowest()`,
			wantErr:  true,
		},
		{
			name:     "query function with side effects",
			builtins: []string{`foo = () => {sideEffect() return 1}`},
			program:  `foo()`,
			want: []values.Value{
				values.NewInt(0),
				values.NewInt(1),
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := dependenciestest.Default().Inject(context.Background())
			r := repl.New(ctx, dependenciestest.Default())
			if _, err := r.Eval(prelude); err != nil {
				t.Fatalf("unable to evaluate prelude: %s", err)
			}

			for _, builtin := range tc.builtins {
				if _, err := r.Eval(builtin); err != nil {
					t.Fatal("evaluation of builtin failed: ", err)
				}
			}

			sideEffects, err := r.Eval(tc.program)
			if err != nil && !tc.wantErr {
				t.Fatal("program evaluation failed: ", err)
			} else if err == nil && tc.wantErr {
				t.Fatal("expected to error during program evaluation")
			}

			if tc.want != nil {
				if want, got := tc.want, getSideEffectsValues(sideEffects); !cmp.Equal(want, got, semantictest.CmpOptions...) {
					t.Fatalf("unexpected side effect values -want/+got: \n%s", cmp.Diff(want, got, semantictest.CmpOptions...))
				}
			}
		})
	}
}

// TestInterpreter_MultipleEval tests that multiple calls to `Eval` to the same interpreter behave as expected.
func TestInterpreter_MultipleEval(t *testing.T) {
	type scriptWithSideEffects struct {
		script      string
		sideEffects []interpreter.SideEffect
	}

	testCases := []struct {
		name  string
		lines []scriptWithSideEffects
	}{
		{
			name: "1 expression statement",
			lines: []scriptWithSideEffects{
				{
					script: `1+1`,
					sideEffects: []interpreter.SideEffect{
						{
							Value: values.NewInt(2),
							Node: &semantic.ExpressionStatement{
								Expression: &semantic.BinaryExpression{
									Left:     &semantic.IntegerLiteral{Value: 1},
									Operator: ast.AdditionOperator,
									Right:    &semantic.IntegerLiteral{Value: 1},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "more expression statements",
			lines: []scriptWithSideEffects{
				{
					script: `1+1`,
					sideEffects: []interpreter.SideEffect{
						{
							Value: values.NewInt(2),
							Node: &semantic.ExpressionStatement{
								Expression: &semantic.BinaryExpression{
									Left:     &semantic.IntegerLiteral{Value: 1},
									Operator: ast.AdditionOperator,
									Right:    &semantic.IntegerLiteral{Value: 1},
								},
							},
						},
					},
				},
				{
					script:      `foo = () => {sideEffect() return 1}`,
					sideEffects: []interpreter.SideEffect{}, // no side effect expected.
				},
				{
					script: `foo()`, // 2 side effects: the function call and the statement expression.
					sideEffects: []interpreter.SideEffect{
						{
							Value: values.NewInt(0),
							Node: &semantic.CallExpression{
								Callee: &semantic.MemberExpression{
									Object:   &semantic.IdentifierExpression{Name: semantic.NewSymbol("testutil")},
									Property: semantic.NewSymbol("yield@testutil"),
								},
								Arguments: &semantic.ObjectExpression{Properties: []*semantic.Property{}},
								Pipe:      &semantic.IntegerLiteral{Value: 0},
							},
						},
						{
							Value: values.NewInt(1),
							Node: &semantic.ExpressionStatement{
								Expression: &semantic.CallExpression{
									Callee:    &semantic.IdentifierExpression{Name: semantic.NewSymbol("foo")},
									Arguments: &semantic.ObjectExpression{Properties: []*semantic.Property{}},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := dependenciestest.Default().Inject(context.Background())
			r := repl.New(ctx, dependenciestest.Default())

			if _, err := r.Eval(prelude); err != nil {
				t.Fatalf("unable to evaluate prelude: %s", err)
			}

			for _, line := range tc.lines {
				if ses, err := r.Eval(line.script); err != nil {
					t.Fatal("evaluation of builtin failed: ", err)
				} else {
					if !cmp.Equal(line.sideEffects, ses, semantictest.CmpOptions...) {
						t.Fatalf("unexpected side effect values -want/+got: \n%s", cmp.Diff(line.sideEffects, ses, semantictest.CmpOptions...))
					}
				}
			}
		})
	}
}

func TestResolver(t *testing.T) {
	testcases := []struct {
		env  string
		fn   string
		want string
	}{
		{
			env:  "x = 42",
			fn:   "f = (r) => r + x",
			want: "(r) => r + 42",
		},
		{
			env:  "v = {env: 42}",
			fn:   "f = (r) => r + v.env",
			want: "(r) => r + 42",
		},
		{
			env:  `option v = {env: "acc"}`,
			fn:   "f = (r) => r.env == v.env",
			want: `(r) => r.env == "acc"`,
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run("", func(t *testing.T) {
			src := tc.env + "\n" + tc.fn

			// Evaluate script with a function definition.
			ctx := dependenciestest.Default().Inject(context.Background())
			_, scope, err := runtime.Eval(ctx, src)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			fn, ok := scope.Lookup("f")
			if !ok {
				t.Fatalf("could not lookup function definition")
			}

			resolver, ok := fn.Function().(interpreter.Resolver)
			if !ok {
				t.Fatalf("function is not resolvable")
			}

			got, err := resolver.Resolve()
			if err != nil {
				t.Fatalf("could not resolve function: %s", err)
			}

			want := executetest.FunctionExpression(t, tc.want)
			if !cmp.Equal(want, got, semantictest.CmpOptions...) {
				t.Errorf("unexpected resoved function: -want/+got\n%s", cmp.Diff(want, got, semantictest.CmpOptions...))
			}
		})
	}
}

func getSideEffectsValues(ses []interpreter.SideEffect) []values.Value {
	vs := make([]values.Value, len(ses))
	for i, se := range ses {
		vs[i] = se.Value
	}
	return vs
}

func TestStack(t *testing.T) {
	src := `from(bucket: "telegraf") |> range(start: -5m) |> aggregateWindow(every: 1m, fn: mean)`
	ctx := dependenciestest.Default().Inject(context.Background())
	sideEffects, _, err := runtime.Eval(ctx, src)
	if err != nil {
		t.Fatal(err)
	} else if len(sideEffects) != 1 {
		t.Fatalf("expected 1 side effect, got %d", len(sideEffects))
	}

	to, ok := sideEffects[0].Value.(*flux.TableObject)
	if !ok {
		t.Fatal("expected side effect to be a table object")
	}

	// The position of window is dependent on the position of the function
	// call in the universe.flux file. If the universe.flux file changes,
	// the corresponding start and end should be changed too.
	want := []interpreter.StackEntry{
		{
			FunctionName: "window",
			Location: ast.SourceLocation{
				File:   "universe.flux",
				Start:  ast.Position{Line: 302, Column: 12},
				End:    ast.Position{Line: 302, Column: 51},
				Source: `window(every: inf, timeColumn: timeDst)`,
			},
		},
		{
			FunctionName: "aggregateWindow",
			Location: ast.SourceLocation{
				Start:  ast.Position{Line: 1, Column: 50},
				End:    ast.Position{Line: 1, Column: 86},
				Source: `aggregateWindow(every: 1m, fn: mean)`,
			},
		},
	}
	got := to.Source.Stack
	if !cmp.Equal(want, got) {
		t.Fatalf("unexpected stack -want/+got:\n%s", cmp.Diff(want, got))
	}
}
