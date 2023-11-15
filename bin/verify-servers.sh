#!/bin/bash
set -euoE pipefail

servers=$(yq eval '.spec.clusters[].nodes[].bmcAddress' "$1" |awk -F '//' '{print "https://"$2}')

source ./redfish-actions/sushy.sh
for s in "${servers[@]}";do
  media_status "$s"
done
