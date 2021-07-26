.PHONY: all clean help

SRC := $(wildcard *.go)

all: 		# build all
	pre-commit install
	gofmt -w -s $(SRC)
	go test -cover -failfast -timeout 2s

clean:		# clean-up the environment
	@find . -name '*.swp' -delete

help:		# show this message
	@printf "Usage: make [OPTION]\n"
	@printf "\n"
	@perl -nle 'print $$& if m{^[\w-]+:.*?#.*$$}' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?#"} {printf "    %-18s %s\n", $$1, $$2}'

doc:		# show the document in local
	godoc -server=localhost:8080 hello.go
