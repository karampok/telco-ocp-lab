#! /usr/bin/env bash
set -euoE pipefail

PULL_SECRET=${PULL_SECRET:-/root/.pull-secret.json}

oc adm release extract --registry-config "${PULL_SECRET}" \
   --command=openshift-install --to /usr/local/bin/ "${OCP_RELEASE}"
openshift-install version

name=${1:-sno} #mno,sno,5gc
folder=${folder:-"/tmp/${name}"}
cp -r "${name}"-template "${folder}"

sed -i "s|PULLSECRET|$(jq '.' -c "$PULL_SECRET")|g" "${folder}"/install-config.yaml
sed -i "s|SSHKEY|$(cat ~/.ssh/authorized_keys)|g" "${folder}"/install-config.yaml

openshift-install agent create image --log-level info --dir "${folder}"
exit 0

source "${HOME}/sushy.sh"
while IFS= read -r node; do
  power_off "$node"
  media_eject "$node"
  media_insert "$node" "${HTTP_SERVER:-http://10.10.20.200:9000}"/"${name}"/agent.x86_64.iso
  boot_once "$node"
  power_on "$node"
done <"${folder}/bmc-hosts"

mkdir -p ~/.kube && cp "${folder}"/auth/kubeconfig ~/.kube/config
openshift-install agent wait-for install-complete --log-level info --dir /share/"${name}"
