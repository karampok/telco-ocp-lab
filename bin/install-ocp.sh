#! /usr/bin/env bash
set -euoE pipefail

name=${1}   #hub,mno
TS=$(awk -F . '{print $1;}' /proc/uptime)
folder="$(pwd)/deployments/${name}-$TS"

mkdir -p "$(pwd)"/deployments/ && cp -r "${name}"-template "${folder}"

# shellcheck disable=1090
source "${folder}"/envs ||true

PULL_SECRET=$(jq '.' -c "${PULL_SECRET_PATH}") #one liner
SSHPUBKEY=$(cat "$SSHPUBKEYFILE")
sed -i "s|SSHPUBKEY|$SSHPUBKEY|" "${folder}"/install-config.yaml
sed -i "s/PULLSECRET/$PULL_SECRET/g" "${folder}"/install-config.yaml

cp -r "${folder}" "${folder}"/archive 2>/dev/null || true

echo "Creating discovery ISO"
openshift-install agent create image --log-level info  --dir "${folder}"
rm -rf /isos/agent-"${name}".iso || true
mv "${folder}"/agent.x86_64.iso  /isos/agent-"${name}".iso
chcon -t httpd_sys_content_t /isos/agent-"${name}".iso || true
#scp ./isos/agent.x86_64.iso infra@192.168.18.10:/var/www/html/iso/ 2>&1

echo "Booting the servers"
# shellcheck disable=1090
source ./redfish-actions/"${BMC_TYPE:-sushy}".sh  # BMC_TYPE=hpe|dell|sushy
for node in $(cat "${folder}"/bmc-hosts);
do
  power_off "$node"
  media_eject "$node"
  media_insert "$node" "${HTTP_SERVER}"/agent-"${name}".iso # e.g. HTTP_SERVER=https://10.9.10.71/isos/
  boot_once "$node"
  power_on "$node"
done

export KUBECONFIG="${folder}"/auth/kubeconfig
openshift-install agent wait-for install-complete --log-level info --dir "${folder}"
