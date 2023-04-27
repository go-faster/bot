all: generate test build

test:
	@./go.test.sh

coverage:
	@./go.coverage.sh

generate:
	go generate
	go generate ./...

build:
	CGO_ENABLED=0 go build ./cmd/bot

check_generated: generate
	git diff --exit-code

forward_psql:
	kubectl -n faster port-forward svc/postgresql 15432:5432

.PHONY: check_generated coverage test generate build forward_psql
