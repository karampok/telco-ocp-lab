RELEASE ?= quay.io/openshift-release-dev/ocp-release:4.22.2-x86_64
NAME ?= sno

.PHONY: iso fetch-bin deploy clean

fetch-bin:
	@openshift-install version 2>/dev/null | grep -q '$(RELEASE)' || \
		oc adm release extract --registry-config "$(HOME)/.pull-secret.json" \
			--command=openshift-install --to $(HOME)/.local/bin/ "$(RELEASE)"

deploy:
	mkdir -p .libvirt/bmh{1,2,3,4,5,6,7,8,9} share
	PUBLICIP=$$(ip --json route get 8.8.8.8 | jq -r '.[].prefsrc') clab deploy --topo topo.yaml

clean:
	sudo -E clab destroy --topo topo.yaml 2>/dev/null || true
	sudo rm -rf .libvirt clab-vlab

iso:
	cp -r $(NAME)-template ./share/$(NAME)
	PS=$$(cat $(HOME)/.pull-secret.json) yq eval -i '.pullSecret = strenv(PS)' ./share/$(NAME)/install-config.yaml
	SK=$$(cat $(HOME)/.ssh/authorized_keys) yq eval -i '.sshKey = strenv(SK)' ./share/$(NAME)/install-config.yaml
	openshift-install agent create image --log-level info --dir ./share/$(NAME)
