#! /usr/bin/env bash
set -euoE pipefail

openshift-install version

cp -r sno-template /share/hub
#read -p -r # tree before,after

PULL_SECRET=$(jq '.' -c "${PULL_SECRET_PATH:-.pull-secret.json}") #one liner
sed -i "s/PULLSECRET/$PULL_SECRET/g" /share/hub/install-config.yaml

openshift-install agent create image --log-level info  --dir /share/hub

#read -p -r # discovery ISO created // nmstatectl required
#read -p -r  # python3 -m http.server 9000 -d /share

node="https://192.168.100.100:8000/redfish/v1/Systems/11111111-1111-1111-1111-111111111115"
source ./redfish-actions/sushy.sh
power_off "$node"
media_eject "$node"
media_insert "$node" http://10.10.20.200:9000/hub/agent.x86_64.iso
boot_once "$node" #dummy in sushy
power_on "$node"


mkdir -p ~/.kube && cp /share/hub/auth/kubeconfig ~/.kube/config
cat << EOF
export KUBECONFIG=/share/hub/auth/kubeconfig
openshift-install agent wait-for install-complete --log-level info --dir /share/hub
EOF
