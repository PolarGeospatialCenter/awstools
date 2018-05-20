.PHONY: test deps all

test: deps
	go test -cover ./pkg/...

all: test

vendor: Gopkg.lock
	dep ensure -vendor-only

deps: vendor
