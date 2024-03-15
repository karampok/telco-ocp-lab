#! /usr/bin/env bash
set -euoE pipefail

PULL_SECRET=${PULL_SECRET:-.pull-secret.json}
OCP_RELEASE=${1:-"quay.io/openshift-release-dev/ocp-release:4.15.3-x86_64"}
# registry.ci.openshift.org/ocp/release@sha256:56fa5020a0bd9e31547ee5de38d3e6065d6df9aa7df21422c9335d634ef5edf5

# 1. official repo
# https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/openshift-install-linux.tar.gz

# 2. official latest dev
# https://amd64.ocp.releases.ci.openshift.org/releasestream/4-dev-preview/release/4.12.0-ec.4

oc adm release extract --registry-config "${PULL_SECRET}" \
  --command=openshift-install --to "/usr/local/bin/" "$OCP_RELEASE"

oc adm release extract --registry-config "${PULL_SECRET}" \
  --command=oc --to "/usr/local/bin/" "$OCP_RELEASE"

openshift-install version
