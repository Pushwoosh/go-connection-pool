all: lint test

.PHONY: lint
lint:
ifeq (, $(shell which golangci-lint))
	$(error "No golangci-lint in $(PATH). Install it from https://github.com/golangci/golangci-lint")
endif
	golangci-lint run


.PHONY: test
test:
	go test ./...