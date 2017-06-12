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
	go test -v -cover -timeout=1m $(PACKAGES) 

testrace:
	go test -v -race -cover -timeout=1m $(PACKAGES)

bench:
	go test -run=NONE -bench=. -benchmem $(PACKAGES)

gen:
	go generate $(PACKAGES)

plugin-backend:
	mkdir -p ./bin
	go build -o ./bin/$(NAME)-$(PLUGIN) ./plugin/backend/$(PLUGIN)/plugin

plugin-server:
	mkdir -p ./bin
	go build -o ./bin/$(NAME)-$(PLUGIN) ./plugin/server/$(PLUGIN)/plugin

plugins:
	make plugin-backend PLUGIN=mongo
	make plugin-backend PLUGIN=redis
	make plugin-server PLUGIN=http
	make plugin-server PLUGIN=nsq
