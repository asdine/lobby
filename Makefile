NAME := lobby

.PHONY: all $(NAME) deps install test testrace bench gen plugin-backend plugin-server plugins

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

plugin-backend:
	mkdir -p ./bin
	go build -o ./bin/$(NAME)-$(PLUGIN) ./plugin/backend/$(PLUGIN)/plugin

plugin-server:
	mkdir -p ./bin
	go build -o ./bin/$(NAME)-$(PLUGIN) ./plugin/server/$(PLUGIN)/plugin

plugins:
	make plugin-backend PLUGIN=mongo
	make plugin-backend PLUGIN=redis
	make plugin-server PLUGIN=nsq
