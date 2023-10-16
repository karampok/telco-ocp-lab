#!/usr/bin/env bash
set -euo pipefail

# script to validate that Manifests are correct it execute the Kustomize
# plugins that will be executed later on ArgoCD it allows to check it before
# pushing changes to be used on our CI/CD

# author: Jose Gato Luis <jgato@redhat.com>

Color_Off='\033[0m'
BGreen='\033[1;32m'       # Green
BYellow='\033[1;33m'      # Yellow

VALIDATE_SRC=$1
ZTP_SITE_GENERATOR_IMG=registry.redhat.io/openshift4/ztp-site-generate-rhel8:v4.13.1
PRE_VALIDATE_ERROR_LOG="/tmp/pre-validate-error-${RANDOM}.log"
echo "Log details in: ${PRE_VALIDATE_ERROR_LOG}"

echo -e "${BYellow}# yamllint *.yaml in ${VALIDATE_SRC} ${Color_Off}"
for f in $(git ls-files "${VALIDATE_SRC}"/'*.yaml')
do
  yamllint "$f" -d relaxed --no-warnings
done

# without is 5sec faster
# if oc get clusterversion 1>/dev/null 2>&1; then
#   ZTP_SITE_GENERATOR_IMG=$(oc -n openshift-gitops get argocd openshift-gitops -o jsonpath='{.spec.repo.initContainers[0].image}')
# fi

get_plugins()
{
    echo -e "Validating with ztp-site-generator: ${ZTP_SITE_GENERATOR_IMG}"
    if [ -d "$1" ]; then
      return
    fi
    mkdir -p "$1"
    podman create --name pgtool ${ZTP_SITE_GENERATOR_IMG}
    podman cp pgtool:/kustomize/plugin/ran.openshift.io "$1" && podman rm pgtool
}

echo -e "${BYellow}# ZTP Manifests in kustomization.yaml ${Color_Off}"
KUSTOMIZE_PLUGIN_HOME=$(git rev-parse --show-toplevel)/.ztp-kustomize-plugin/"${ZTP_SITE_GENERATOR_IMG##*/}"
get_plugins "$KUSTOMIZE_PLUGIN_HOME"
export KUSTOMIZE_PLUGIN_HOME
kustomize build "${VALIDATE_SRC}" --enable-alpha-plugins 1>.ztp-kustomize-plugin/out.yaml 2>${PRE_VALIDATE_ERROR_LOG}


#podman run --rm --log-driver=none -v "$(pwd)":/resources:Z,U ${ZTP_SITE_GENERATOR_IMG} \
#  generator config "${VALIDATE_SRC}"
#  generator install sites/n62-ocp2.yaml


# TODO
#cat .ztp-kustomize-plugin/out.yaml |  sed -E -e's/(namespace:)(.+)/\1 default\n/g' | oc apply --dry-run=server -f - &>> ${PRE_VALIDATE_ERROR_LOG}
