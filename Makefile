.PHONY: all clean help lint

GENERATE_SRC := type_string.go typehint_string.go
SRC := $(wildcard *.go) $(wildcard */*.go) ${GENERATE_SRC}
BIN := $(subst .go,,$(wildcard examples/*.go))

all: $(BIN)	# build all
	@pre-commit install

clean:		# clean-up the environment
	@find . -name '*.swp' -delete
	rm -f $(BIN) $(GENERATE_SRC)

help:		# show this message
	@printf "Usage: make [OPTION]\n"
	@printf "\n"
	@perl -nle 'print $$& if m{^[\w-]+:.*?#.*$$}' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?#"} {printf "    %-18s %s\n", $$1, $$2}'

doc:		# show the document in local
	godoc -server=localhost:8080 hello.go

$(GENERATE_SRC):
	@go get golang.org/x/tools/cmd/stringer
	@PATH=$$PATH:$(shell go env GOPATH)/bin/ go generate

$(BIN): lint

lint: $(SRC)
	@gofmt -w -s $^
	go test -cover -failfast -timeout 2s

%: %.go
	go build -ldflags="-s -w" -o $@ $<
