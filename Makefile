# Makefile configuration
.DEFAULT_GOAL := help
.PHONY: help release

fmt: ## Golang code formatting tool
	@printf "\033[0;32mRunning formatting tool\033[0m\n"
	@gofmt -s -w .

deps: ## Download required dependencies
	@printf "\033[0;32mInstalling dependencies\033[0m\n"
	go get ./...

release: ## Builds release
	@printf "\n\033[0;32mBuilding binaries\033[0m\n"
	@rm -rf release/
	@mkdir -p release/
	GOOS="linux" go build -o release/checker-linux checker/main.go
	GOOS="linux" go build -o release/target-linux target/main.go
	GOOS="darwin" go build -o release/checker-darwin checker/main.go
	GOOS="darwin" go build -o release/target-darwin target/main.go
	@printf "\n\033[0;32mDone\033[0m\n"

help:
	@grep --extended-regexp '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
