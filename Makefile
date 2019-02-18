commit := $(shell git rev-parse --short HEAD)
ldflags := -ldflags "-X main.version=$(commit)"
application := cmd/crawler/main.go
binary := dist/crawler
hostname ?= integralist.co.uk
subdomains ?= "www,"

# additional make command properties that are mapped to cli flags...
#
# httponly: will attempt to normalize requests to HTTP protocol
# example: make run httonly=-httponly
#
# json: will output only the final json (so there could be a long pause before you see anything)
# example: make run json=-json
#
# dot: will output only a dot formatted file for use with graphviz dot command (similar delay to -json)
# example: dot=-dot

test:
	go test -v -failfast ./...

run:
	@go run $(ldflags) $(application) -hostname $(hostname) -subdomains $(subdomains) $(httponly) $(json) ${dot}

build:
	go build $(ldflags) -o $(binary) $(application)

clean:
	@rm $(binary) &> /dev/null || true
