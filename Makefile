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
	go build -o $(NAME)-$(PLUGIN) ./plugin/backend/$(PLUGIN)

plugin-server:
	go build -o $(NAME)-$(PLUGIN) ./plugin/server/$(PLUGIN) 

plugins:
	make plugin-backend PLUGIN=bolt
	make plugin-server PLUGIN=http