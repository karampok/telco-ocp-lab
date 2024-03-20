
.PHONY: build clean clean doc

infra:
	./telco-ocp-lab --setup -a -auto-timeout 0s

clean:
	./telco-ocp-lab --clean -a -auto-timeout 0s

doc:
	./telco-ocp-lab --setup --dry-run -a -i -auto-timeout 0s --no-color
