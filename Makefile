NAME      := brazier
PACKAGES  := $(shell glide novendor)

.PHONY: gen

gen:
	go generate $(PACKAGES)
