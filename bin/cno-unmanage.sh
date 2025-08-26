#!/usr/bin/env bash

set -e

usage() {
  printf "\n$0 <namespace> <override>\n    Patch the Cluster Version Operator to not manage the specified component\n"
}

deployment() {
  ns="${1:-openshift-network-operator}"
  name="${2:-network-operator}"
  value="${3:-true}"

  [[ -z "$ns" ]] && {
    echo "error: undefined namespace"
    usage
    exit 1
  }
  [[ -z "$name" ]] && {
    echo "error: undefined deployment override name"
    usage
    exit 1
  }

  # retrieve the current overrides list and filter using jq
  local index=$(kubectl get clusterversion version -o json | jq -r '.spec.overrides // [] | to_entries | map(select(.value.name=="'$name'" and .value.namespace=="'$ns'")) | if length > 0 then .[0].key else "-" end')
  local op="replace"
  if [ "$index" = "-" ]; then
    op="add"
    overrides_exists=$(kubectl get clusterversion version -o json | jq -r '.spec.overrides // empty')
    if [ -z "$overrides_exists" ]; then
      # if .spec.overrides is not present, create it with the new entry
      kubectl patch clusterversion version --type='json' -p "$(
        cat <<-EOF
[
  {
    "op": "add",
    "path": "/spec/overrides",
    "value": [
      {
        "group": "apps",
        "kind": "Deployment",
        "namespace": "$ns",
        "name": "$name",
        "unmanaged": $value
      }
    ]
  }
]
EOF
      )"
      return
    fi

  fi

  kubectl patch clusterversion version --type='json' -p "$(
    cat <<-EOF
[
  {
    "op": "$op",
    "path": "/spec/overrides/$index",
    "value": {
      "group": "apps",
      "kind": "Deployment",
      "namespace": "$ns",
      "name": "$name",
      "unmanaged": $value
    }
  }
]
EOF
  )"
}

crd() {
  name="$1"
  value="${2:-true}"
  echo "TODO: fix implementation to use jsonpatch"
  exit 1

  [[ -z "$name" ]] && {
    echo "error: undefined CRD override name"
    usage
    exit 1
  }

  kubectl patch clusterversion version --type='merge' -p "$(
    cat <<-EOF
spec:
  overrides:
  - group: config.openshift.io
    kind: CustomResourceDefinition
    name: $name
    namespace: ""
    unmanaged: $value
EOF
  )"
}

main() {
  if [[ "$1" == "crd" ]]; then
    shift
    crd "$@"
  else
    deployment "$@"
  fi
}

main "$@"
