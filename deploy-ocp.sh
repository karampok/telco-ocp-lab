#! /usr/bin/env bash
set -euoE pipefail

openshift-install version

name=${1}   #hpe183,sno,5gc
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
  media_insert "$node" "${HTTP_SERVER:-http://192.168.100.200:9000}"/"${name}"/agent-"${name}".iso
  boot_once "$node"
  power_on "$node"
done

mkdir -p ~/.kube && cp "${folder}"/auth/kubeconfig ~/.kube/config
cat << EOF
openshift-install agent wait-for install-complete --log-level info --dir /share/hub
EOF
