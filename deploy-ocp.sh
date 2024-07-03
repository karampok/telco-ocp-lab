#! /usr/bin/env bash
set -euoE pipefail

PULL_SECRET=${PULL_SECRET:-/root/.pull-secret.json}

OCP_RELEASE=${1:-"quay.io/openshift-release-dev/ocp-release:4.16.1-x86_64"}
oc adm release extract --registry-config "${PULL_SECRET}" \
  --command=openshift-install --to "/usr/local/bin/" "$OCP_RELEASE"
openshift-install version

name=${1:-mno}  #mno,sno,5gc
folder=${folder:-"/share/${name}"}
cp -r "${name}"-template "${folder}"

sed -i "s/PULLSECRET/$(jq '.' -c "$PULL_SECRET")/g" "${folder}"/install-config.yaml

openshift-install agent create image --log-level info --dir "${folder}"

source "${folder}"/redfish-actions/sushy.sh
while IFS= read -r node; do
  power_off "$node"
  media_eject "$node"
  media_insert "$node" "${HTTP_SERVER:-http://192.168.100.200:9000}"/"${name}"/agent.x86_64.iso
  boot_once "$node"
  power_on "$node"
done < "${folder}/bmc-hosts"

mkdir -p ~/.kube && cp "${folder}"/auth/kubeconfig ~/.kube/config
openshift-install agent wait-for install-complete --log-level info --dir /share/${name}

cat << EOF
oc patch network.operator cluster --type=merge --patch-file day1/network-operator-patch.yaml
oc patch OperatorHub cluster --type merge --patch-file day1/operatorhub-patch.yaml
# Local registry is needed
# https://docs.openshift.com/container-platform/4.15/registry/configuring_registry_storage/configuring-registry-storage-baremetal.html
oc patch configs.imageregistry cluster --type=merge --patch-file day1/image-registry-patch.yaml
# twice to remove any topologySpreadConstraints: []
oc patch configs.imageregistry cluster --type=merge --patch-file day1/image-registry-patch.yaml
EOF
