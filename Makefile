plugin_name = selectel
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)

clean:
	rm -rf bin/$(plugin_name)

build: clean
	GOGC=off go build -o bin/$(plugin_name) cmd/main.go

install: build
	cp bin/$(plugin_name) /usr/local/bin/docker-machine-driver-$(plugin_name)

vet:
	@echo "go vet ."
	@go vet $$(go list ./... | grep -v vendor/) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

fmt:
	gofmt -w $(GOFMT_FILES)

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

errcheck:
	@sh -c "'$(CURDIR)/scripts/errcheck.sh'"

imports:
	goimports -w $(GOFMT_FILES)

importscheck:
	@sh -c "'$(CURDIR)/scripts/goimportscheck.sh'"

lintcheck:
	@sh -c "'$(CURDIR)/scripts/golintcheck.sh'"

vendor-status:
	dep status
