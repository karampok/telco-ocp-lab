#!/bin/bash
set -euoE pipefail

REPO=${GITREPO:-karampok/telco-ocp-lab.git}
KEY=${GITKEY_FILE:-.github-argo}

if [[ ! -f $KEY ]]
then
  ssh-keygen -q -t ed25519 -C "github-repo" -f "$KEY" <<< $'\ny' >/dev/null 2>&1
fi

# https://github.com/argoproj/argo-cd/blob/master/docs/operator-manual/argocd-repositories.yaml
oc create -n openshift-gitops secret generic ztp-spokes-repo \
     --from-file=sshPrivateKey="$KEY" \
     --from-literal=type=git \
     --from-literal=url=git@github.com:"$REPO" \
     --from-literal=insecure=true

oc label -n openshift-gitops secret ztp-spokes-repo argocd.argoproj.io/secret-type=repository

echo "sshPrivateKey has been added into ArgoCD repository."
echo
echo "Then go to https://github.com/${REPO%".git"}/settings/keys"
echo "Add a new deploy key"
echo "Title: argocd"
echo "Key: $(cat "$KEY.pub")"
echo
# or
# argocd login openshift-gitops-server-openshift-gitops.apps.x.com --sso
#
# oc get secret -n openshift-gitops openshift-gitops-cluster -o jsonpath="{.data.admin\.password}" |base64 -d
# argocd login openshift-gitops-server-openshift-gitops.apps.hub.eric.vlab --username admin --insecure
#
# argocd repo add git@github.com:$REPO --ssh-private-key-path /tmp/github-ed25519 --insecure --insecure-ignore-host-key"
# argocd repo add https://github.com/argoproj/argocd-example-apps --username <username> --password <password>
# argocd repo list
#
# argocd app get openshift-gitops/hub-provision --refresh && argocd app sync openshift-gitops/hub-provision --prune
# argocd app get openshift-gitops/clusters --refresh && argocd app sync openshift-gitops/clusters --prune
# argocd app get openshift-gitops/policies --refresh && argocd app sync openshift-gitops/policies --prune
