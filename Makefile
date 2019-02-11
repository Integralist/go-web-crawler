commit := $(shell git rev-parse --short HEAD)
ldflags := -ldflags "-X main.version=$(commit)"
application := cmd/crawler/main.go
binary := dist/crawler
hostname ?= integralist.co.uk
subdomains ?= "www,"

# note: httponly=-httponly will result in only HTTP protocol being valid

test:
	go test -v -failfast ./...

run:
	go run $(ldflags) $(application) -hostname $(hostname) -subdomains $(subdomains) $(httponly)

build:
	go build $(ldflags) -o $(binary) $(application)

clean:
	@rm $(binary) &> /dev/null || true
