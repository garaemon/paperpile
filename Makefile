BINARY := paperpile-cli
GO := go

.PHONY: build clean lint test

build:
	$(GO) build -o $(BINARY) .

clean:
	rm -f $(BINARY)

lint:
	$(GO) vet ./...

test:
	$(GO) test ./...
