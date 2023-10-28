# esazoth

esazoth recives reindex task ID and wait completed and returns the recommended document sync batch period.

## Install

```sh
$ go install github.com/po3rin/esazoth@latest
```

## Usage

```sh
result=$(curl -X POST \
  'http://localhost:9200/_reindex?wait_for_completion=false' \
  -H 'Content-Type: application/json' \
  -d '{
    "source": {
        "index": "test-src"
    },
    "dest": {
        "index": "test-dist"
    }
}' | jq -r .task esazoth )

# reindex took 3 days ...

echo $result #3
```

Result of esazoth can be used to fill in document differences that occur after reindexing.
If you are updating documents using k8s cronjob, execute as follows.

```sh
$ start_days_ago=$result
$ kubectl create job --from=cronjob/$cronjob_name -n $namespace $job_name --dry-run=client -o "json" \
| jq --arg start_days_ago "$start_days_ago" '( .spec.template.spec.containers[0].env[] | select(.name == "START_DAYS_AGO") ).value |= $start_days_ago' \
| kubectl apply -f -
```
