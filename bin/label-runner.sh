#!/bin/bash
set -euoE pipefail

# Get the runner's ID
RUNNER_ID=$(curl -sS -H "Authorization: Bearer $GH_TOKEN" \
              -H "Accept: application/vnd.github.v3+json" \
              "https://api.github.com/repos/karampok/telco-ocp-lab/actions/runners" | \
              jq --arg RN "$RUNNER" -r '.runners[] | select(.name == $RN) | .id')

curl -X POST \
     -s --noproxy "*"  -w "%{http_code} %{url_effective}\\n" -L --globoff \
     -H "Authorization: Bearer $GH_TOKEN" \
     -H "Accept: application/vnd.github.v3+json" \
     -H "Content-Type: application/json" \
     -d '{"labels": ["'"$1"'"]}' \
     "https://api.github.com/repos/karampok/telco-ocp-lab/actions/runners/${RUNNER_ID}/labels"

# curl -X DELETE \
#     -s --noproxy "*"  -w "%{http_code} %{url_effective}\\n" -L --globoff \
#     -H "Authorization: Bearer $GH_TOKEN" \
#     -H "Accept: application/vnd.github.v3+json" \
#     "https://api.github.com/repos/karampok/telco-ocp-lab/actions/runners/${RUNNER_ID}/labels/ready"
