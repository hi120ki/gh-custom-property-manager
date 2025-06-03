all: help

help: ## Print this help message
	@grep -E '^[a-zA-Z._-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: apply
apply: ## Apply changes
	GITHUB_TOKEN=$$(gh auth token) go run main.go apply --config property/property-a.yaml --config property/property-b.yaml

.PHONY: plan
plan: ## Plan changes
	GITHUB_TOKEN=$$(gh auth token) go run main.go plan --config property/property-a.yaml --config property/property-b.yaml

.PHONY: test
test: ## Run tests
	go test -v ./... -cover
