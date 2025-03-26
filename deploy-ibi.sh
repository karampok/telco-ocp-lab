#! /usr/bin/env bash
set -euoE pipefail

date

PULL_SECRET=${PULL_SECRET:-${HOME}/.pull-secret.json}
OCP_RELEASE=${OCP_RELEASE:-"quay.io/openshift-release-dev/ocp-release:4.18.6-x86_64"}

oc adm release extract --registry-config "${PULL_SECRET}" \
  --command=openshift-install --to "${HOME}/.local/bin/" "$OCP_RELEASE"
openshift-install version

date

name=${1:-sno} #mno,sno,5gc
folder=${folder:-"/share/${name}"} && mkdir -p "$folder"
cp -r "${name}"-template/* "${folder}"

sed -i "s/PULLSECRET/$(jq '.' -c "$PULL_SECRET")/g" "${folder}"/image-based-installation-config.yaml
sed -i "s/PULLSECRET/$(jq '.' -c "$PULL_SECRET")/g" "${folder}"/install-config.yaml

# needs ImageBasedInstallationConfig
openshift-install image-based create image --log-level info --dir "${folder}"

date

source "${folder}"/redfish-actions/sushy.sh
while IFS= read -r node; do
  power_off "$node"
  media_eject "$node"
  media_insert "$node" "${HTTP_SERVER:-http://192.168.100.200:9000}"/"${name}"/rhcos-ibi.iso
  boot_once "$node"
  power_on "$node"
done <"${folder}/bmc-hosts"

echo "Waiting for part1 - installation"
until ssh -o StrictHostKeyChecking=no -i ~/.ssh/github-actions core@10.10.10.225 'sudo journalctl -lu install-rhcos-and-restore-seed.service | grep Finished' 2>/dev/null; do
  echo -n .
  sleep 30
done

echo
date

# needs ImageBasedConfig and install-config
openshift-install image-based create config-image --dir "${folder}"

date

while IFS= read -r node; do
  power_off "$node"
  media_eject "$node"
  media_insert "$node" "${HTTP_SERVER:-http://192.168.100.200:9000}"/"${name}"/imagebasedconfig.iso
  # boot_once "$node" this is not needed, does not boot
  power_on "$node"
done <"${folder}/bmc-hosts"

mkdir -p ~/.kube && cp "${folder}"/auth/kubeconfig ~/.kube/config
echo "Waiting for part2 - configuration"
until [ "$(oc get clusterversion -o jsonpath='{.items[*].status.conditions[?(@.type=="Available")].status}' 2>/dev/null)" == "True" ]; do
  echo -n .
  sleep 30
done

echo
date

oc get clusterversion
