setup:
	eskeeper < eskeeper.yml
	curl -X PUT 'localhost:9200/test-src/_doc/1' \
	  -H 'Content-Type: application/json' \
	  -d '{"id": 1, "title": "title", "body": "body"}'
