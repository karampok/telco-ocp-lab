#! /usr/bin/env bash
set -xeuoE pipefail

AI_URL=${1:-"10.10.10.116:8090"}
url="http://$AI_URL/api/assisted-install/v2/clusters/"
until ret=$(curl -s -o /dev/null -w "%{http_code}" "$url"); do
  if [ "$ret" -eq 200 ]; then
      break
  fi
  sleep 30;
done

CLUSTER_ID=$(curl "$url" | jq '.[0].id' | tr -d '"')
curl -X PUT -H "Content-Type: application/json" \
         -d '{"host-validation-ids": "[\"belongs-to-majority-group\"]"}' \
         "$url"/"$CLUSTER_ID"/ignored-validations
