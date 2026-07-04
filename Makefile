GO ?= go
GOFLAGS ?= -mod=vendor
export GOPROXY=off

.PHONY: help build test fmt verify verify-solution verify-attacks cover
help: ## list targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  %-18s %s\n",$$1,$$2}'
build: ## compile
	$(GO) build $(GOFLAGS) ./...
test: ## run unit + functional tests (app + functional only; exploit tests run via harness)
	$(GO) test $(GOFLAGS) -count=1 ./app/... ./tests/functional/...
verify: ## grade the BASE app (expect every instance reward 0)
	python3 harness/run_eval.py --all --base
verify-solution: ## grade with each golden applied (expect every instance reward 1.0)
	python3 harness/run_eval.py --all --solution
verify-attacks: ## prove each adversarial patch scores reward 0
	python3 harness/run_eval.py --attacks
cover: ## coverage report (repo CI only; never the reward)
	$(GO) test $(GOFLAGS) -count=1 -coverprofile=cover.out ./app/... && $(GO) tool cover -func=cover.out | tail -1
fmt: ## gofmt + go vet check (repo CI only; never the reward)
	@test -z "$$(gofmt -l app/ tests/)" || { echo "gofmt needed on:"; gofmt -l app/ tests/; exit 1; }
	$(GO) vet $(GOFLAGS) ./app/... ./tests/...
