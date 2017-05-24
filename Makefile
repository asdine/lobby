NAME      := lobby
PACKAGES  := $(shell glide novendor)

.PHONY: all $(NAME) deps install test testrace bench gen plugin

all: $(NAME)

$(NAME):
	go install ./cmd/$@

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

plugin:
	go build -o $(NAME)-$(PLUGIN) ./plugin/backend/$(PLUGIN) 
