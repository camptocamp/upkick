DEPS = $(wildcard */*.go)
VERSION = $(shell git describe --always --dirty)

all: test upkick upkick.1

upkick: main.go $(DEPS)
	CGO_ENABLED=0 GOOS=linux \
	  go build -a \
		  -ldflags="-X main.version=$(VERSION)" \
	    -installsuffix cgo -o $@ $<
	strip $@

upkick.1: upkick
	./upkick -m > $@

lint:
	@ go get -v github.com/golang/lint/golint
	@for file in $$(git ls-files '*.go' | grep -v '_workspace/'); do \
		export output="$$(golint $${file} | grep -v 'type name will be used as docker.DockerInfo')"; \
		[ -n "$${output}" ] && echo "$${output}" && export status=1; \
	done; \
	exit $${status:-0}

vet: main.go
	go vet $<

imports: main.go
	goimports -d $<

test: lint vet imports
	go test -v ./...

coverage:
	rm -rf *.out
	go test -coverprofile=coverage.out

clean:
	rm -f upkick upkick.1

.PHONY: all lint vet imports test coverage clean
