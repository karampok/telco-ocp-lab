#! /usr/bin/env bash
set -euoE pipefail

openshift-install version

name=${1:-mno}   #mno,sno,5gc
folder=${folder:-"/share/${name}"}
cp -r "${name}"-template "${folder}"

PULL_SECRET=$(jq '.' -c "${PULL_SECRET_PATH:-.pull-secret.json}") #one liner
sed -i "s/PULLSECRET/$PULL_SECRET/g" "${folder}"/install-config.yaml

openshift-install agent create image --log-level info  --dir "${folder}"

source ./redfish-actions/sushy.sh
for node in $(cat "${folder}"/bmc-hosts);
do
  power_off "$node"
  media_eject "$node"
  media_insert "$node" "${HTTP_SERVER:-http://10.10.20.200:9000}"/"${name}"/agent.x86_64.iso
  boot_once "$node"
  power_on "$node"
done

mkdir -p ~/.kube && cp "${folder}"/auth/kubeconfig ~/.kube/config
#openshift-install agent wait-for install-complete --log-level info --dir /share/${name}
