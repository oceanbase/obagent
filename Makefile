include Makefile.common

.PHONY: all test clean build monagent

default: clean fmt build

build: build-debug

build-debug: set-debug-flags monagent

build-release: set-release-flags monagent

set-debug-flags:
	@echo Build with debug flags
	$(eval LDFLAGS += $(LDFLAGS_DEBUG))

set-release-flags:
	@echo Build with release flags
	$(eval LDFLAGS += $(LDFLAGS_RELEASE))

monagent:
	$(GOBUILD) $(GO_RACE_FLAG) -ldflags '$(MONAGENT_LDFLAGS)' -o bin/monagent cmd/monagent/main.go

test:
	$(GOTEST) $(GOTEST_PACKAGES)

fmt:
	@gofmt -s -w $(filter-out , $(GOFILES))

fmt-check:
	@if [ -z "$(UNFMT_FILES)" ]; then \
		echo "gofmt check passed"; \
		exit 0; \
    else \
    	echo "gofmt check failed, not formatted files:"; \
    	echo "$(UNFMT_FILES)" | tr -s " " "\n"; \
    	exit 1; \
    fi

tidy:
	$(GO) mod tidy

vet:
	go vet $$(go list ./...)

clean:
	rm -rf $(GOCOVERAGE_FILE)
	rm -rf tests/mock/*
	rm -rf bin/monagent
	$(GO) clean -i ./...
