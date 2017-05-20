NAME      := lobby
PACKAGES  := $(shell glide novendor)

.PHONY: deps install test testrace bench gen

deps:
	glide up

install:
	glide install

test:
	go test -v -cover $(PACKAGES)

testrace:
	go test -v -race -cover $(PACKAGES)

bench:
	go test -run=NONE -bench=. -benchmem $(PACKAGES)

gen:
	go generate $(PACKAGES)
