APPS = $(notdir $(wildcard ./apps/*))

generate-proto:
	cd service/apps/app/proto && buf generate
.PHONY: generate-proto

build-apps:
	rm -rf ./dist/apps
	@for app in $(APPS); do \
		echo "Building $$app"; \
		mkdir -p ./dist/apps/$$app; \
		go build -o ./dist/apps/$$app/app ./apps/$$app; \
	done
.PHONY: build-apps

### LOCAL
local.build: build-apps
	go build -o dist/autopus ./cmd/cli
.PHONY: local.build

local.example:
	go run cmd/cli/main.go example
.PHONY: local.example

local.test.integration:
	go test -tags=integration -v ./test/integration/...

local.start-service:
	docker-compose up -d --build
.PHONY: local.start-service

local.stop-service:
	docker-compose down
.PHONY: local.stop-service

test.unit: test.generate-fixtures
	go test -v ./...
.PHONY: test.unit

test.generate-fixtures:
	go build -o ./test/fixtures/apps/hello/app ./test/fixtures/hello
.PHONY: test.generate-fixtures

test.start-service: test.stop-service
	docker-compose -f docker-compose.integration.yml up -d --build
	docker ps
	docker logs autopus
.PHONY: test.start-service

test.stop-service:
	docker-compose -f docker-compose.integration.yml down
.PHONY: test.stop-service

test.integration: test.start-service
	PORT=8088 go test -tags=integration -v ./test/integration/...
	docker logs autopus
.PHONY: test.integration

build.ui:
	cd ui && yarn build

build.ui.dev:
	cd ui && yarn build-dev

deploy: build-apps
	fly deploy