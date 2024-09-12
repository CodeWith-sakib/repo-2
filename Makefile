.PHONY: all build test test-race vet clean loc

all: build test

build:
	go build ./...

test:
	go test -v ./...

test-race:
	go test -race ./...

vet:
	go vet ./...

loc:
	@python3 -c "import generator.loc as l; res=l.get_production_loc('.'); print(f'Production LOC: {res["code"]}')"

clean:
	go clean
	rm -rf bin/
