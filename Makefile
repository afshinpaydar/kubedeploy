.PHONY: help

help: ## Helpful commands
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.DEFAULT_GOAL := build

build: ## Build the docker image in local and build kubectl-deploy locally
	@echo 'build docker image'
	docker build -t kubectl-deploy:latest .
	go mod download
	mkdir -p ./build
	go build -o ./build/kubectl-deploy .

clean: ## Clear docker artifacts from local machine
	@echo 'clear project build folder'
	rm -rf ./build

install: build ## Install kubectl-deploy
	go install kubectl-deploy



