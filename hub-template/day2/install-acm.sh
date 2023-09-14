#!/bin/bash
set -euoE pipefail

# we need to leverage https://github.com/redhat-cop/gitops-catalog/
#oc apply -k https://github.com/redhat-cop/gitops-catalog/advanced-cluster-management/operator/overlays/release-2.8
#oc apply -k https://github.com/redhat-cop/gitops-catalog/lvm-storage/operator/overlays/4.13-stable
oc apply -f 00-advanced-cluster-management-operator.yaml
oc apply -f 00-lvm-storage.yaml
oc apply -f 00-openshift-gitops.yaml
oc apply -f 00-talm.yaml
until oc get MultiClusterHub 2>/dev/null; do sleep 30; done

oc apply -f 10-provisioning.yaml
oc apply -f 10-lvm-instance.yaml
oc apply -f 10-acm-instance.yaml
until oc -n openshift-storage get lvmcluster 2>/dev/null; do sleep 30; done
oc -n openshift-storage wait lvmcluster lvmcluster --for=jsonpath='{.status.state}'=Ready --timeout=1200s
oc -n open-cluster-management wait MultiClusterHub multiclusterhub --for=jsonpath='{.status.phase}'=Running --timeout=1200s

oc apply -f 20-agent-service-config.yaml
