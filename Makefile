COVERAGE_THRESH_PCT=0
.PHONY: prep
prep: clean
	@echo "Making sure Go is installed"
	@go version
	@echo "Making sure golangci-lint is installed"
	@golangci-lint version

.PHONY: test
test: lint
	go test --coverprofile cover.out ./...
	overcover --coverprofile cover.out ./... --threshold $(COVERAGE_THRESH_PCT)

.PHONY: lint
lint:
	golangci-lint run --timeout 5m

.PHONY: build
build:
	go build

#List targets in makefile
.PHONY: list
list:
	@$(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$'
