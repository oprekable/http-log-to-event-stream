.PHONY: docker-pull-hurl
docker-pull-hurl:
	@docker pull ghcr.io/orange-opensource/hurl:latest

.PHONY: docker-compose-build
docker-compose-build:
	@docker compose -p http-log-to-event-stream build

.PHONY: docker-compose-down
docker-compose-down:
	@docker compose -p http-log-to-event-stream down; \
	rm -rf ./.docker/redis/data

.PHONY: docker-compose-up
docker-compose-up:
	@docker compose -p http-log-to-event-stream up -d

.PHONY: docker-cleanup
docker-cleanup:
	@docker system prune --all; \
     docker volume prune -f; \
     docker images prune -a;

.PHONY: call-apis
call-apis:
	@docker run --rm -v ./.docker/hurl:/data -w /data ghcr.io/orange-opensource/hurl:latest --test \
	/data/http-OK-api-protected-logged.hurl \
	/data/http-non-OK-api-protected-logged.hurl \
	/data/api-not-protected-logged.hurl \
	/data/api-not-protected-delay.hurl

.PHONY: install-loadtest-tools
install-loadtest-tools:
	@go install github.com/tsenart/vegeta@latest

.PHONY: load-test
load-test:
	@jq -ncM 'while(true; .+1) | {method: "GET", url: ("http://host.docker.internal:8000/" + (if (. == null) then 0 else . end | if ((. % 2) == 0) then "one" else "two" end) +"/template?total=" + (. | tostring) + "&response=" + (if (. == null) then 0 else . end | if ((. % 5) == 0) then 500 else 200 end | tostring)), header: {"Authorization": ["Bearer any-random-token"]} }' \
	| vegeta attack -rate=60/s -duration=300s -lazy -format=json | tee results.bin | vegeta report

.PHONY: load-test-delay
load-test-delay:
	@jq -ncM 'while(true; .+1) | {method: "GET", url: ("http://host.docker.internal:8000/three/template?total=" + (. | tostring)) }' \
	| vegeta attack -rate=60/s -duration=300s -lazy -format=json | tee results.bin | vegeta report

.PHONY: load-test-log-of-unknown-upstream
load-test-log-of-unknown-upstream:
	@jq -ncM 'while(true; .+1) | {method: "GET", url: ("http://host.docker.internal:8000/four/template?total=" + (. | tostring)) }' \
	| vegeta attack -rate=60/s -duration=300s -lazy -format=json | tee results.bin | vegeta report

.PHONY: check-redis-url
check-redis-url:
	@echo XRANGE mockoon-one - + | redis-cli -h 127.0.0.1 -p 16379 | grep "service" | jq -s '.[].request.url'

