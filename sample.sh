#!/usr/bin/env bash

# required
# * esazoth
# * curl
# * jq
# * kubectl 

# env
# * ESAZOTH_ES_USER
# * ESAZOTH_ES_PASS
# * SLACK_CHANNEL_WEBHOOK_URL

if [ $# -ne 2 ]; then
  echo "3 arguments is required" 1>&2
  exit 1
fi

namespace=$1
cronjob_name=$2
job_name="{$cronjob_name}-esazoth"
buffer=3

result=$(curl -X POST \
  -u "${ESAZOTH_ES_USER}:${ESAZOTH_ES_PASSWORD}" \
  'http://localhost:9200/_reindex?wait_for_completion=false' \
  -H 'Content-Type: application/json' \
  -d '{
    "source": {
        "index": "test-src"
    },
    "dest": {
        "index": "test-dist"
    }
}' | jq .task | tr -d '"' | go run ./... )

if [ $? -ne 0 ]; then
  msg="Reindexの監視が失敗しました。ログを確認してください。${result}"
  echo $msg
  # slack_body='{"username": "esazoth", "attachments": [{"mrkdwn_in": "text", "text": "'${msg}'", "color": "danger"}]}'
  # curl -s -X POST -H 'Content-type: application/json' -d "$slack_body" "${SLACK_CHANNEL_WEBHOOK_URL}"
  exit 1
fi

start_days_ago=$(($result+$buffer))

# msg="Reindexが完了しました。Reindex中に更新された差分を埋めるために${start_days_ago}日前から${namespace}/${cronjob_name}をJobとして実行します。"
# slack_body='{"username": "esazoth", "attachments": [{"mrkdwn_in": "text", "text": "'${msg}'", "color": "good"}]}'
# curl -s -X POST -H 'Content-type: application/json' -d "$slack_body" "${SLACK_CHANNEL_WEBHOOK_URL}"

kubectl create job --from=cronjob/$cronjob_name -n $namespace $job_name --dry-run=client -o "json" \
| jq --arg start_days_ago "$start_days_ago" '( .spec.template.spec.containers[0].env[] | select(.name == "START_DAYS_AGO") ).value |= $start_days_ago' \
| kubectl apply -f -
