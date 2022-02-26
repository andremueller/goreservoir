.PHONY: all
all:
	cd cmd/reservoir && go build

.PHONY: update
update:
	go get -u ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: test
test:
	go test ./...
