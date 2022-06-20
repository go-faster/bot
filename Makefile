test:
	@./go.test.sh
.PHONY: test

coverage:
	@./go.coverage.sh
.PHONY: coverage

generate:
	go generate
	go generate ./...
.PHONY: generate

build:
	CGO_ENABLED=0 go build ./cmd/bot

check_generated: generate
	git diff --exit-code
.PHONY: check_generated
