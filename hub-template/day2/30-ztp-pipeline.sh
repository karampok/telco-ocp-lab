#! /usr/bin/env bash
set -euoE pipefail

out=/tmp/out
mkdir -p "$out"


podman run --log-driver=none --rm --authfile=/workdir/.pull-secret.json \
  registry.redhat.io/openshift4/ztp-site-generate-rhel8:v4.12 extract /home/ztp/argocd/deployment --tar | tar x -C "$out"

# or
# git clone -b release-4.12 git@github.com:openshift-kni/cnf-features-deploy.git /tmp/out
# cp -r /tmp/ou2/ztp/gitops-subscriptions/argocd/deployment "$out" 

oc patch argocd openshift-gitops -n openshift-gitops --type=merge \
  --patch-file "$out"/argocd-openshift-gitops-patch.json

# increase memory/cpu request/limits
# kustomizeBuildOptions: --enable-alpha-plugins
# add init-container

REPO="git@github.com:karampok/telco-ocp-lab.git"
echo "---
apiVersion: v1
kind: Namespace
metadata:
  name: clusters-sub
---
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: clusters
  namespace: openshift-gitops
spec:
  destination:
    server: https://kubernetes.default.svc
    namespace: clusters-sub
  project: ztp-app-project
  source:
    path: ztp-spokes/sites
    repoURL: $REPO
    targetRevision: gitops
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true" > "$out"/clusters-app.yaml

echo "---
apiVersion: v1
kind: Namespace
metadata:
  name: policies-sub
---
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: policies
  namespace: openshift-gitops
spec:
  destination:
    server: https://kubernetes.default.svc
    namespace: policies-sub
  project: policy-app-project
  source:
    path: ztp-spokes/policies
    repoURL: $REPO
    targetRevision: gitops
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true" > "$out"/policies-app.yaml

oc apply -k "$out"/
