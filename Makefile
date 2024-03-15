GO_VERSION := 1.22.1

.PHONY: build clean clean doc

build:
	@echo "Building Go binary..."
	@podman run --rm -v $(PWD):/src:z -w /src golang:$(GO_VERSION) go build

infra:
	./telco-ocp-lab --setup -a -auto-timeout 0s

clean:
	./telco-ocp-lab --clean -a -auto-timeout 0s

doc:
	./telco-ocp-lab --setup --dry-run -a -i -auto-timeout 0s --no-color
