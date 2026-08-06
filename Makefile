.PHONY = list build-api build-link clear-tasks ngrok redis-cli redis-server run-api run-link test

API_SRC := ./cmd/api/main.go
API_BIN := api
LINK_SRC := ./cmd/authorize/main.go
LINK_BIN := link
REDIS_CONFIG := ./redis.config

list:
	@echo "Make targets:\n"
	@LC_ALL=C $(MAKE) -pRrq -f $(firstword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/(^|\n)# Files(\n|$$)/,/(^|\n)# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | grep -E -v -e '^[^[:alnum:]]' -e '^$@$$'

build-api: ${API_SRC}
	@echo "Building api bin..."
	go build ${API_SRC} -o ${API_BIN}

build-link: ${LINK_SRC}
	@echo "Building link bin..."
	go build ${LINK_SRC} -o ${LINK_BIN}

clear-tasks:
	@echo "Clearing task message queue..."
	@rm dump.rdb && rm -rf appendonlydir

ngrok:
	@echo "Running ngrok..."
	ngrok http 8080

redis-cli: ${REDIS_CLI}
	@echo "Starting redis cli..."
	@redis-cli

redis-server: ${REDIS_CONFIG}
	@echo "Starting redis server..."
	redis-server ${REDIS_CONFIG}

run-api: ${API_SRC}
	@echo "Running api server..."
	go run ${API_SRC}

run-link: ${LINK_SRC}
	@echo "Sending link..."
	go run ${LINK_SRC}

test:
	@echo "Running tests..."
	go test ./...
