#! /usr/bin/env bash
set -euoE pipefail

openshift-install version
#read -p -r # https://mirror.openshift.com/pub/openshift-v4/clients/ocp/4.13.9/

cp -r hub-template /share/hub
#read -p -r # tree before,after

PULL_SECRET=$(jq '.' -c "${PULL_SECRET_PATH:-.pull-secret}") #one liner
SSHPUBKEY=$(cat "${SSHPUBKEYFILE:-.id-rsa.pub}")
sed -i "s|SSHPUBKEY|$SSHPUBKEY|" /share/hub/install-config.yaml
sed -i "s/PULLSECRET/$PULL_SECRET/g" /share/hub/install-config.yaml


openshift-install agent create image --log-level info  --dir ./share/hub

#read -p -r # discovery ISO created // nmstatectl required
#read -p -r  # python3 -m http.server 9000 -d /share


# "Mount and booting the ISO in the servers using RedFish"
node="https://192.168.100.100:8000/redfish/v1/Systems/11111111-1111-1111-1111-111111111115"
source ./redfish-actions/sushy.sh
power_off "$node"
media_eject "$node"
media_insert "$node" http://192.168.100.200:9000/agent-5gc.iso
boot_once "$node"
power_on "$node"

read -p -r
# kcli console  5gc-m0; kcli console 5gc-m1; kcli console 5gc-m2

cat << EOF
export KUBECONFIG=./deployments/5gc/auth/kubeconfig
openshift-install agent wait-for install-complete --log-level info --dir "./deployments/5gc"
EOF

# talk about install-config, agent-config
# talk about day0 operation
# talk about dual stack VIPs // br-ex
