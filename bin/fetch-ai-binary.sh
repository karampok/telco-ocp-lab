#! /usr/bin/env bash
set -euoE pipefail

# 1. official repo
# https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/openshift-install-linux.tar.gz

# 2. official latest dev
# https://amd64.ocp.releases.ci.openshift.org/releasestream/4-dev-preview/release/4.12.0-ec.4
# or https://amd64.ocp.releases.ci.openshift.org/graph

#export REGISTRY_AUTH_FILE=.pull-secret.json
PULL_SECRET_PATH=${PULL_SECRET_PATH:-.pull-secret.json}

oc adm release extract --registry-config "${PULL_SECRET_PATH}" --command=openshift-install --to "/usr/local/bin/" \
   quay.io/openshift-release-dev/ocp-release:4.15.3-x86_64

oc adm release extract --registry-config "${PULL_SECRET_PATH}" --command=oc --to "/usr/local/bin/" \
   quay.io/openshift-release-dev/ocp-release:4.15.3-x86_64

#oc adm release extract --registry-config .pull-secret.json --command=openshift-install --to "$(pwd)/bin" \
# registry.ci.openshift.org/ocp/release@sha256:56fa5020a0bd9e31547ee5de38d3e6065d6df9aa7df21422c9335d634ef5edf5

# 3 directly from source
# https://github.com/openshift-metal3/dev-scripts/blob/master/agent/release/billi-release.sh
# ./billi-release.sh c1005941e quay.io/openshift-release-dev/ocp-release:4.12.0-ec.4-x86_64
# where commit in https://github.com/openshift/installer

openshift-install version


# NOTE: payload on the binary is different from the source code of the binary
# NOTE: assisted-service commit under  AGENT-INSTALLER-API-SERVER
