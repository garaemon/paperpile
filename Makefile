BINARY := paperpile
GO := go

.PHONY: build clean lint test coverage fmt-check

build:
	$(GO) build -o $(BINARY) .

clean:
	rm -f $(BINARY) coverage.out

lint:
	$(GO) vet ./...
	staticcheck ./...

test:
	$(GO) test -race ./...

coverage:
	$(GO) test -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GO) tool cover -func=coverage.out

fmt-check:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "Files not formatted:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi
