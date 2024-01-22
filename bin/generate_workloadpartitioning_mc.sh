#!/bin/bash

CPUSET=$1; shift

if [ -z "$CPUSET" ]
then
    echo "Run the script like ./generate_workloadpartitioning_config.sh <CPUSET>, for example ./generate_workloadpartitioning_config.sh 0-3,16-19"
    exit 1
fi

DATA1=$(echo "[crio.runtime.workloads.management]
activation_annotation = \"target.workload.openshift.io/management\"
annotation_prefix = \"resources.workload.openshift.io\"
resources = { \"cpushares\" = 0, \"cpuset\" = \"$CPUSET\" }" | base64 -w0)

DATA2=$(echo "{\"management\": {\"cpuset\": \"$CPUSET\"}}" | base64 -w0)

echo "---
apiVersion: machineconfiguration.openshift.io/v1
kind: MachineConfig
metadata:
  labels:
    machineconfiguration.openshift.io/role: ht100gb
  name: 02-ht100gb-workload-partitioning
spec:
  config:
    ignition:
      version: 3.2.0
    storage:
      files:
      - contents:
          source: data:text/plain;charset=utf-8;base64,$DATA1
        mode: 420
        overwrite: true
        path: /etc/crio/crio.conf.d/01-workload-partitioning
        user:
          name: root
      - contents:
          source: data:text/plain;charset=utf-8;base64,$DATA2
        mode: 420
        overwrite: true
        path: /etc/kubernetes/openshift-workload-pinning
        user:
          name: root"

