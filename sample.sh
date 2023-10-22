#!/usr/bin/env bash

if [ $# -ne 2 ]; then
  echo "3 arguments is required" 1>&2
  exit 1
fi

namespace=$1
cronjob_name=$2
job_name="{$cronjob_name}-esazoth"

start_days_ago=$(curl -X POST \
  'http://localhost:9200/_reindex?wait_for_completion=false' \
  -H 'Authorization: Basic XXXXXXX' \
  -H 'Content-Type: application/json' \
  -d '{
    "source": {
        "index": "test-src"
    },
    "dest": {
        "index": "test-dist"
    }
}' | jq .task | tr -d '"' | go run ./cmd/esazoth/...)

echo $start_days_ago

kubectl create job --from=cronjob/$cronjob_name -n $namespace $job_name --dry-run=client -o "json" \
| jq --arg start_days_ago "$start_days_ago" '( .spec.template.spec.containers[0].env[] | select(.name == "START_DAYS_AGO") ).value |= $start_days_ago' \
| kubectl apply -f -