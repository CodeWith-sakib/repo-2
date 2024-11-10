GO ?= /opt/homebrew/bin/go
export GOROOT := /opt/homebrew/Cellar/go/1.26.1/libexec
unexport GOBIN

.PHONY: all build test test-race vet clean loc

all: build test

build:
	$(GO) build -buildvcs=false ./...

test:
	$(GO) test -v ./...

test-race:
	$(GO) test -race ./...

vet:
	$(GO) vet ./...

loc:
	@python3 -c "import generator.loc as l; res=l.get_production_loc('.'); print(f'Production LOC: {res[\"code\"]}')"

clean:
	$(GO) clean
	rm -rf bin/
