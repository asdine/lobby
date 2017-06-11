NAME      := lobby
PACKAGES  := $(shell glide novendor)

.PHONY: all $(NAME) deps install test testrace bench gen plugin-backend plugin-server plugins

all: $(NAME) plugins

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

plugin-backend:
	mkdir -p ./bin
	go build -o ./bin/$(NAME)-$(PLUGIN) ./builtin/backend/$(PLUGIN)/plugin

plugin-server:
	mkdir -p ./bin
	go build -o ./bin/$(NAME)-$(PLUGIN) ./builtin/server/$(PLUGIN)/plugin

plugins:
	make plugin-backend PLUGIN=mongo
	make plugin-server PLUGIN=http
	make plugin-server PLUGIN=nsq
