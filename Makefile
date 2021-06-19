
.PHONY: all
all:
	go build ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: test
test:
	go test ./...
