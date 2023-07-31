include Makefile.common

.PHONY: pre-build build bindata mgragent monagent agentd agentctl mockgen rpm buildsucc

default: clean fmt pre-build build

pre-build: bindata

pre-test: mockgen

bindata: get
	go-bindata -o bindata/bindata.go -pkg bindata assets/...

mockgen:
	mockgen -source=lib/http/http.go -destination=tests/mock/lib_http_http_mock.go -package mock
	mockgen -source=lib/system/process.go -destination=tests/mock/lib_system_process_mock.go -package mock
	mockgen -source=lib/system/disk.go -destination=tests/mock/lib_system_disk_mock.go -package mock
	mockgen -source=lib/system/system.go -destination=tests/mock/lib_system_system_mock.go -package mock
	mockgen -source=lib/shell/shell.go -destination=tests/mock/lib_shell_shell_mock.go -package mock
	mockgen -source=lib/shell/command.go -destination=tests/mock/lib_shell_command_mock.go -package mock
	mockgen -source=lib/shellf/shellf.go -destination=tests/mock/lib_shellf_shellf_mock.go -package mock
	mockgen -source=lib/file/file.go -destination=tests/mock/lib_file_file_mock.go -package mock
	mockgen -source=lib/pkg/package.go -destination=tests/mock2/lib_pkg_package_mock.go -package mock2

build: build-debug

build-debug: set-debug-flags mgragent monagent agentd agentctl buildsucc

build-release: set-release-flags mgragent monagent agentd agentctl buildsucc

rpm:
	cd ./rpm && RELEASE=`date +%Y%m%d%H%M%S` rpmbuild -bb ./obagent.spec

set-debug-flags:
	@echo Build with debug flags
	$(eval LDFLAGS += $(LDFLAGS_DEBUG))
	$(eval BUILD_FLAG += $(GO_RACE_FLAG))	

set-release-flags:
	@echo Build with release flags
	$(eval LDFLAGS += $(LDFLAGS_RELEASE))

monagent:
	$(GO) build $(BUILD_FLAG) -ldflags '$(MONAGENT_LDFLAGS)' -o bin/ob_monagent cmd/monagent/main.go

mgragent: 
	$(GO) build $(BUILD_FLAG) -ldflags '$(MGRAGENT_LDFLAGS)' -o bin/ob_mgragent cmd/mgragent/main.go

agentctl:
	$(GO) build $(BUILD_FLAG) -ldflags '$(AGENTCTL_LDFLAGS)' -o bin/ob_agentctl cmd/agentctl/main.go

agentd:
	$(GO) build $(BUILD_FLAG) -ldflags '$(AGENTD_LDFLAGS)' -o bin/ob_agentd cmd/agentd/main.go


buildsucc:
	@echo Build obagent successfully!

runmgragent:
	./bin/ob_mgragent --config tests/testdata/mgragent.yaml

runmonagent:
	./bin/ob_monagent --config tests/testdata/monagent.yaml

test: pre-build pre-test
	$(GOTEST) $(GOTEST_PACKAGES)

test-cover-html:
	go tool cover -html=$(GOCOVERAGE_FILE)

test-cover-html-out:
	mkdir -p $(GOCOVERAGE_REPORT)
	go tool cover -html=$(GOCOVERAGE_FILE) -o $(GOCOVERAGE_REPORT)/index.html

test-cover-profile:
	go tool cover -func=$(GOCOVERAGE_FILE)

test-cover-total:
	go tool cover -func=$(GOCOVERAGE_FILE) | tail -1 | awk '{print "total line coverage: " $$3}'

deploy:
	mkdir /home/admin/obagent
	cp -r bin /home/admin/obagent/bin
	cp -r etc /home/admin/obagent/conf
	mkdir -p /home/admin/obagent/{log,run,tmp,backup,pkg_store,task_store,position_store}

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

get:
	$(GO) install github.com/go-bindata/go-bindata/...@v3.1.2+incompatible
	$(GO) install github.com/golang/mock/mockgen@v1.6.0

vet:
	go vet $$(go list ./...)

clean:
	rm -rf $(GOCOVERAGE_FILE)
	rm -rf tests/mock/*
	rm -rf bin/ob_mgragent bin/ob_monagent bin/ob_agentctl bin/ob_agentd
	$(GO) clean -i ./...

init: pre-build pre-test tidy
