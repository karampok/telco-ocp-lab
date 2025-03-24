#! /usr/bin/env bash
set -euoE pipefail

PULL_SECRET=${PULL_SECRET:-${HOME}/.pull-secret.json}
OCP_RELEASE=${OCP_RELEASE:-"quay.io/openshift-release-dev/ocp-release:4.18.6-x86_64"}

oc adm release extract --registry-config "${PULL_SECRET}" \
  --command=openshift-install --to "${HOME}/.local/bin/" "$OCP_RELEASE"
openshift-install version

name=${1:-sno} #mno,sno,5gc
folder=${folder:-"/share/${name}"} && mkdir -p "$folder"
cp -r "${name}"-template/* "${folder}"

sed -i "s/PULLSECRET/$(jq '.' -c "$PULL_SECRET")/g" "${folder}"/image-based-installation-config.yaml
sed -i "s/PULLSECRET/$(jq '.' -c "$PULL_SECRET")/g" "${folder}"/install-config.yaml

#openshift-install agent create image --log-level info --dir "${folder}"
openshift-install image-based create image --log-level info --dir "${folder}"
openshift-install image-based create config-image --dir "${folder}"

source "${folder}"/redfish-actions/sushy.sh
while IFS= read -r node; do
  power_off "$node"
  media_eject "$node"
  media_insert "$node" "${HTTP_SERVER:-http://192.168.100.200:9000}"/"${name}"/rhcos-ibi.iso
  boot_once "$node"
  power_on "$node"
done <"${folder}/bmc-hosts"

# mkdir -p ~/.kube && cp "${folder}"/auth/kubeconfig ~/.kube/config
# openshift-install agent wait-for install-complete --log-level info --dir /share/${name}
#
# cat <<EOF
# oc patch network.operator cluster -p '{"spec":{"defaultNetwork":{"ovnKubernetesConfig":{"gatewayConfig":{"routingViaHost": true}}}}}' --type=merge
# oc patch network.operator cluster --type=merge --patch-file day1/network-operator-patch.yaml
# oc patch OperatorHub cluster --type merge --patch-file day1/operatorhub-patch.yaml
# # Local registry is needed
# # https://docs.openshift.com/container-platform/4.15/registry/configuring_registry_storage/configuring-registry-storage-baremetal.html
# oc patch configs.imageregistry cluster --type=merge --patch-file day1/image-registry-patch.yaml
# # twice to remove any topologySpreadConstraints: []
# oc patch configs.imageregistry cluster --type=merge --patch-file day1/image-registry-patch.yaml
# oc patch OperatorHub cluster --type json -p '[{"op": "add", "path": "/spec/disableAllDefaultSources", "value": true}]'
# oc patch Scheduler cluster --type=merge --patch '{ "spec": { "mastersSchedulable": true } }'
# EOF
