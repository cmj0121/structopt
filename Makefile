.PHONY: all clean help lint

SRC := $(wildcard *.go) $(wildcard */*.go)
BIN := examples/example

all: $(BIN)	# build all
	@pre-commit install

clean:		# clean-up the environment
	@find . -name '*.swp' -delete
	rm -f $(BIN)

help:		# show this message
	@printf "Usage: make [OPTION]\n"
	@printf "\n"
	@perl -nle 'print $$& if m{^[\w-]+:.*?#.*$$}' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?#"} {printf "    %-18s %s\n", $$1, $$2}'

doc:		# show the document in local
	godoc -server=localhost:8080 hello.go

$(BIN): lint

lint:
	gofmt -w -s $(SRC)
	go test -cover -failfast -timeout 2s

%: %.go
	go build -ldflags="-s -w" -o $@ $<
