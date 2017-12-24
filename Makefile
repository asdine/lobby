NAME := lobby

.PHONY: all $(NAME) deps install test testrace bench gen plugin plugins

all: $(NAME)

$(NAME):
	go install ./cmd/$@

deps:
	glide up

install:
	glide install

test:
	go test -v -cover -timeout=1m ./... 

testrace:
	go test -v -race -cover -timeout=2m ./...

bench:
	go test -run=NONE -bench=. -benchmem ./...

gen:
	go generate ./...

plugin:
	mkdir -p ./bin
	go build -o ./bin/$(NAME)-$(PLUGIN) ./plugin/backend/$(PLUGIN)/plugin

plugins:
	make plugin PLUGIN=mongo
	make plugin PLUGIN=redis
	make plugin PLUGIN=nsq
