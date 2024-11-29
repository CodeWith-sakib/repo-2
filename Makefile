export PATH := /opt/homebrew/bin:/opt/homebrew/sbin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin
GO ?= /opt/homebrew/bin/go
export GOROOT := /opt/homebrew/Cellar/go/1.26.1/libexec
unexport GOBIN

.PHONY: all build test test-race vet staticcheck clean loc

all: build test

build:
	$(GO) build -buildvcs=false ./...

test:
	$(GO) test -v ./...

test-race:
	$(GO) test -race ./...

vet:
	$(GO) vet ./...

staticcheck:
	@if command -v staticcheck >/dev/null 2>&1; then staticcheck ./...; else $(GO) vet ./...; fi

loc:
	@python3 -c "import generator.loc as l; res=l.get_production_loc('.'); print(f'Production LOC: {res[\"code\"]}')"

clean:
	$(GO) clean
	rm -rf bin/
