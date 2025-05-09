#!/bin/bash
set -euoE pipefail

repo="karampok/telco-ocp-lab"

runners=$(gh api "repos/${repo}/actions/runners")
# gh api == curl -sS -H "Authorization: Bearer $GITHUB_TOKEN" -H "Accept: application/vnd.github.v3+json" "https://api.github.com/${repo}/actions/runners")

if [ $# -eq 0 ]; then
  echo "$runners" | jq -r '.runners |= sort_by(.name) | .runners[] | "\(.name): \(.labels[3:] | map(.name) | join(" "))"'
  # echo "gh run list --workflow=ztp-compact.yaml -s in_progress -c $(git rev-parse HEAD)"
  # echo "gh run list --workflow=ztp-compact.yaml --json 'workflowDatabaseId' --jq '.[0].workflowDatabaseId'"
  # echo "gh api repos/$repo/actions/runs --paginate -q '.workflow_runs[] | (.id)' | xargs -n1 -I % gh api repos/$repo/actions/runs/% -X DELETE"
  exit 0
fi

RUNNER=$1
shift

RUNNER_ID=$(curl -sS -H "Authorization: Bearer $GITHUB_TOKEN" \
              -H "Accept: application/vnd.github.v3+json" \
              "https://api.github.com/repos/karampok/telco-ocp-lab/actions/runners" | \
              jq --arg RN "$RUNNER" -r '.runners[] | select(.name == $RN) | .id')

add_label() {
  curl -X POST \
      -s --noproxy "*"  -w "%{http_code} $RUNNER +$2 %{url_effective}\\n" -L --globoff -o /dev/null \
      -H "Authorization: Bearer $GITHUB_TOKEN" \
      -H "Accept: application/vnd.github.v3+json" \
      -H "Content-Type: application/json" \
      -d '{"labels": ["'"$2"'"]}' \
      "https://api.github.com/repos/karampok/telco-ocp-lab/actions/runners/${1}/labels"
}

delete_label() {
  curl -X DELETE \
      -s --noproxy "*"  -w "%{http_code} $RUNNER -$2 %{url_effective}\\n" -L --globoff -o /dev/null \
      -H "Authorization: Bearer $GITHUB_TOKEN" \
      -H "Accept: application/vnd.github.v3+json" \
      "https://api.github.com/repos/karampok/telco-ocp-lab/actions/runners/${1}/labels/${2}"
}


while [ "$#" -gt 0 ]; do
    case "$1" in
        +*)
            add_label "$RUNNER_ID" "${1:1}"
            ;;
        -*)
            delete_label "$RUNNER_ID" "${1:1}"
            ;;
        *)
            add_label "$RUNNER_ID" "${1}"
            ;;
    esac
    shift
done
