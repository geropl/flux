# This Makefile encodes the "go generate" prerequisites ensuring that the proper tooling is installed and
# that the generate steps are executed when their prerequeisites files change.

SHELL := /bin/bash

GO_TAGS=
GO_ARGS=-tags '$(GO_TAGS)'

# This invokes a custom package config during Go builds, such that
# the Rust library libflux is built on the fly.
export PKG_CONFIG:=$(shell pwd)/pkg-config.sh

export GOOS=$(shell go env GOOS)
export GO_BUILD=env GO111MODULE=on go build $(GO_ARGS)
export GO_TEST=env GO111MODULE=on go test $(GO_ARGS)
export GO_RUN=env GO111MODULE=on go run $(GO_ARGS)
export GO_TEST_FLAGS=
# Do not add GO111MODULE=on to the call to go generate so it doesn't pollute the environment.
export GO_GENERATE=go generate $(GO_ARGS)
export GO_VET=env GO111MODULE=on go vet $(GO_ARGS)
export CARGO=cargo
export CARGO_ARGS=
export PATH := $(shell pwd)/bin:$(PATH)

define go_deps
	$(shell env GO111MODULE=on go list -f "{{range .GoFiles}} {{$$.Dir}}/{{.}}{{end}}" $(1))
endef

default: build

STDLIB_SOURCES = $(shell find . -name '*.flux')

GENERATED_TARGETS = \
	ast/internal/fbast/ast_generated.go \
	ast/asttest/cmpopts.go \
	stdlib/packages.go \
	internal/fbsemantic/semantic_generated.go \
	internal/fbsemantic/semantic_generated.go \
	internal/feature/flags.go \
	libflux/go/libflux/buildinfo.gen.go \
	$(LIBFLUX_GENERATED_TARGETS)

LIBFLUX_GENERATED_TARGETS = \
	libflux/flux-core/src/ast/flatbuffers/ast_generated.rs \
	libflux/flux-core/src/semantic/flatbuffers/semantic_generated.rs

generate: $(GENERATED_TARGETS)

ast/internal/fbast/ast_generated.go: ast/ast.fbs
	$(GO_GENERATE) ./ast
libflux/flux-core/src/ast/flatbuffers/ast_generated.rs: ast/ast.fbs
	flatc --rust -o libflux/flux-core/src/ast/flatbuffers ast/ast.fbs && rustfmt $@

internal/fbsemantic/semantic_generated.go: internal/fbsemantic/semantic.fbs
	$(GO_GENERATE) ./internal/fbsemantic
libflux/flux-core/src/semantic/flatbuffers/semantic_generated.rs: internal/fbsemantic/semantic.fbs
	flatc --rust -o libflux/flux-core/src/semantic/flatbuffers internal/fbsemantic/semantic.fbs && rustfmt $@
libflux/go/libflux/buildinfo.gen.go: $(LIBFLUX_GENERATED_TARGETS)
	$(GO_GENERATE) ./libflux/go/libflux

# Force a second expansion to happen so the call to go_deps works correctly.
.SECONDEXPANSION:
ast/asttest/cmpopts.go: ast/ast.go ast/asttest/gen.go $$(call go_deps,./internal/cmd/cmpgen)
	$(GO_GENERATE) ./ast/asttest

stdlib/packages.go: $(STDLIB_SOURCES) libflux-go internal/fbsemantic/semantic_generated.go
	$(GO_GENERATE) ./stdlib

internal/feature/flags.go: internal/feature/flags.yml
	$(GO_GENERATE) ./internal/feature

libflux: $(LIBFLUX_GENERATED_TARGETS)
	cd libflux && $(CARGO) build $(CARGO_ARGS)

build: libflux-go
	$(GO_BUILD) ./...

clean:
	rm -rf bin
	cd libflux && $(CARGO) clean && rm -rf pkg
	cd libflux/c && $(MAKE) clean

cleangenerate:
	rm -rf $(GENERATED_TARGETS)

fmt-go:
	go fmt ./...

fmt-rust:
	cd libflux; $(CARGO) fmt

fmt-flux:
	$(GO_RUN) ./cmd/flux/main.go fmt -w ./stdlib

fmt: $(SOURCES_NO_VENDOR) fmt-go fmt-rust fmt-flux

checkfmt:
	./etc/checkfmt.sh

tidy:
	GO111MODULE=on go mod tidy

checktidy:
	./etc/checktidy.sh

checkgenerate:
	./etc/checkgenerate.sh

checkrelease:
	./gotool.sh github.com/goreleaser/goreleaser check

# Run this in two passes to to keep memory usage down. As of this commit,
# running on everything (./...) uses just over 4G of memory. Breaking stdlib
# out keeps memory under 3G.
staticcheck:
	GO111MODULE=on go mod vendor # staticcheck looks in vendor for dependencies.
	GO111MODULE=on ./gotool.sh honnef.co/go/tools/cmd/staticcheck \
		`go list ./... | grep -v '\/flux\/stdlib\>'`
	GO111MODULE=on ./gotool.sh honnef.co/go/tools/cmd/staticcheck ./stdlib/...

test: test-go test-rust test-flux

test-go: libflux-go
	$(GO_TEST) $(GO_TEST_FLAGS) ./...

test-rust:
	cd libflux && $(CARGO) test $(CARGO_ARGS) --all-features && \
	$(CARGO) test --doc && \
	$(CARGO) clippy $(CARGO_ARGS) -- -Dclippy::all

INTEGRATION_WRITE_TESTS=integration_hdb_write_to,integration_sqlite_write_to,integration_vertica_write_to,integration_mssql_write_to,integration_mysql_write_to,integration_mariadb_write_to,integration_pg_write_to
INTEGRATION_READ_TESTS=integration_hdb_read_from_seed,integration_hdb_read_from_nonseed,integration_sqlite_read_from_seed,integration_sqlite_read_from_nonseed,integration_vertica_read_from_seed,integration_vertica_read_from_nonseed,integration_mssql_read_from_seed,integration_mssql_read_from_nonseed,integration_mariadb_read_from_seed,integration_mariadb_read_from_nonseed,integration_mysql_read_from_seed,integration_mysql_read_from_nonseed,integration_pg_read_from_seed,integration_pg_read_from_nonseed
INTEGRATION_TESTS="$(INTEGRATION_WRITE_TESTS),$(INTEGRATION_READ_TESTS)"

test-flux:
	$(GO_RUN) ./cmd/flux test -v --skip $(INTEGRATION_TESTS)

test-flux-integration:
	./etc/spawn-dbs.sh
	$(GO_RUN) ./cmd/flux test -v --test $(INTEGRATION_WRITE_TESTS)
	$(GO_RUN) ./cmd/flux test -v --test $(INTEGRATION_READ_TESTS)

test-race: libflux-go
	$(GO_TEST) -race -count=1 ./...

test-bench: libflux-go
	$(GO_TEST) -run=NONE -bench=. -benchtime=1x ./...
	cd libflux && $(CARGO) test --benches

vet: libflux-go
	$(GO_VET) ./...

bench: libflux-go
	$(GO_TEST) -bench=. -run=^$$ ./...

# This requires ragel 7.0.1.
libflux/flux-core/src/scanner/scanner_generated.rs: libflux/flux-core/src/scanner/scanner.rl
	ragel-rust -I libflux/flux-core/src/scanner -o $@ $<
	rm libflux/flux-core/src/scanner/scanner_generated.ri

# This target generates a file that forces the go libflux wrapper
# to recompile which forces pkg-config to run again.
libflux-go: $(LIBFLUX_GENERATED_TARGETS)
	$(GO_GENERATE) ./libflux/go/libflux

libflux-wasm:
	cd libflux/flux && CC=clang AR=llvm-ar wasm-pack build --scope influxdata --dev

clean-wasm:
	rm -rf libflux/flux/pkg

build-wasm:
	cd libflux/flux && CC=clang AR=llvm-ar wasm-pack build -t nodejs --scope influxdata

publish-wasm: clean-wasm build-wasm
	cd libflux/flux/pkg && npm publish --access public

test-wasm: clean-wasm build-wasm
	cd libflux/flux && CC=clang AR==llvm-ar wasm-pack test --node

test-valgrind: libflux
	cd libflux/c && $(MAKE) test-valgrind

# Build the set of supported cross-compiled binaries
test-release: Dockerfile_build
	docker build -t flux-release -f Dockerfile_build .
	docker run --rm -it -v "$(PWD):/home/builder/src" flux-release /bin/sh -c "\
		cd src/ &&\
	    go build -o /go/bin/pkg-config github.com/influxdata/pkg-config &&\
		./gotool.sh github.com/goreleaser/goreleaser release --rm-dist --snapshot"


bin/flux:
	$(GO_BUILD) -o ./bin/flux ./internal/cmd/flux


libflux/target/release/fluxc: libflux
	cd libflux && $(CARGO) build $(CARGO_ARGS) --release --bin fluxc

libflux/target/release/fluxdoc: libflux
	cd libflux && $(CARGO) build $(CARGO_ARGS) --features=doc --release --bin fluxdoc

fluxdocs: $(STDLIB_SOURCES) libflux/target/release/fluxc libflux/target/release/fluxdoc bin/flux
	FLUXC=./libflux/target/release/fluxc FLUXDOC=./libflux/target/release/fluxdoc ./etc/gen_docs.sh

checkdocs: $(STDLIB_SOURCES) libflux/target/release/fluxc libflux/target/release/fluxdoc bin/flux
	FLUXC=./libflux/target/release/fluxc FLUXDOC=./libflux/target/release/fluxdoc ./etc/checkdocs.sh

# This list is sorted for easy inspection
.PHONY: bench \
	build \
	build-wasm \
	checkdocs \
	checkfmt \
	checkgenerate \
	checkrelease \
	checktidy \
	clean \
	clean-wasm \
	cleangenerate \
	default \
	fluxdocs \
	fmt \
	generate \
	libflux \
	libflux-go \
	libflux-wasm \
	publish-wasm \
	release \
	staticcheck \
	test \
	test-bench \
	test-go \
	test-race \
	test-rust \
	test-valgrind \
	tidy \
	vet
