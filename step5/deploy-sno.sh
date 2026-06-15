#! /usr/bin/env bash
set -euoE pipefail

oc adm release extract --registry-config "${PULL_SECRET}" \
  --command=openshift-install --to "${HOME}/.local/bin/" "$OCP_RELEASE"

openshift-install version
cp -r sno /share/sno
sed -i "s|PULLSECRET|$(jq '.' -c "$PULL_SECRET")|g" /share/sno/install-config.yaml
openshift-install agent create image --log-level info --dir  /share/sno
