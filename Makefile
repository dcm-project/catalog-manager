BINARY_NAME := catalog-manager
# COMPOSE: compose command. Set to override; otherwise auto-detect podman-compose or docker-compose.
COMPOSE ?= $(shell command -v podman-compose >/dev/null 2>&1 && echo podman-compose || \
	(command -v docker-compose >/dev/null 2>&1 && echo docker-compose || \
	(echo "docker compose")))


build:
	go build -o bin/$(BINARY_NAME) ./cmd/$(BINARY_NAME)

run:
	DB_TYPE=sqlite DB_NAME=/tmp/catalog.db go run ./cmd/$(BINARY_NAME)

clean:
	rm -rf bin/

fmt:
	gofmt -s -w .

vet:
	go vet ./...

lint:
	golangci-lint run ./...

test:
	go run github.com/onsi/ginkgo/v2/ginkgo -r --randomize-all --fail-on-pending --skip-package=test/subsystem

tidy:
	go mod tidy

generate-types:
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen \
		--config=api/v1alpha1/types.gen.cfg \
		-o api/v1alpha1/types.gen.go \
		api/v1alpha1/openapi.yaml

generate-spec:
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen \
		--config=api/v1alpha1/spec.gen.cfg \
		-o api/v1alpha1/spec.gen.go \
		api/v1alpha1/openapi.yaml

generate-server:
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen \
		--config=internal/api/server/server.gen.cfg \
		-o internal/api/server/server.gen.go \
		api/v1alpha1/openapi.yaml

generate-client:
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen \
		--config=pkg/client/client.gen.cfg \
		-o pkg/client/client.gen.go \
		api/v1alpha1/openapi.yaml

generate-api: generate-types generate-spec generate-server generate-client generate-service-types

check-generate-api: generate-api
	git diff --exit-code api/ internal/api/server/ pkg/client/ || \
		(echo "Generated files out of sync. Run 'make generate-api'." && exit 1)

# Check AEP compliance
check-aep:
	spectral lint --fail-severity=warn ./api/v1alpha1/openapi.yaml


# Generate Go types for service specifications (VM, Container, Database, Cluster)
generate-service-types:
	@echo "Generating common types..."
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen \
		--config=api/v1alpha1/servicetypes/types.gen.cfg \
		-o api/v1alpha1/servicetypes/types.gen.go \
		api/v1alpha1/servicetypes/common.yaml
	@echo "Generating VM types..."
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen \
		--config=api/v1alpha1/servicetypes/vm/spec.gen.cfg \
		--import-mapping=../common.yaml:github.com/dcm-project/catalog-manager/api/v1alpha1/servicetypes \
		-o api/v1alpha1/servicetypes/vm/types.gen.go \
		api/v1alpha1/servicetypes/vm/spec.yaml
	@echo "Generating Container types..."
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen \
		--config=api/v1alpha1/servicetypes/container/spec.gen.cfg \
		--import-mapping=../common.yaml:github.com/dcm-project/catalog-manager/api/v1alpha1/servicetypes \
		-o api/v1alpha1/servicetypes/container/types.gen.go \
		api/v1alpha1/servicetypes/container/spec.yaml
	@echo "Generating Database types..."
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen \
		--config=api/v1alpha1/servicetypes/database/spec.gen.cfg \
		--import-mapping=../common.yaml:github.com/dcm-project/catalog-manager/api/v1alpha1/servicetypes \
		-o api/v1alpha1/servicetypes/database/types.gen.go \
		api/v1alpha1/servicetypes/database/spec.yaml
	@echo "Generating Cluster types..."
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen \
		--config=api/v1alpha1/servicetypes/cluster/spec.gen.cfg \
		--import-mapping=../common.yaml:github.com/dcm-project/catalog-manager/api/v1alpha1/servicetypes \
		-o api/v1alpha1/servicetypes/cluster/types.gen.go \
		api/v1alpha1/servicetypes/cluster/spec.yaml
	@echo "Generating Three-Tier App Demo types..."
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen \
		--config=api/v1alpha1/servicetypes/three_tier_app_demo/spec.gen.cfg \
		--import-mapping=../common.yaml:github.com/dcm-project/catalog-manager/api/v1alpha1/servicetypes \
		-o api/v1alpha1/servicetypes/three_tier_app_demo/types.gen.go \
		api/v1alpha1/servicetypes/three_tier_app_demo/spec.yaml
	@echo "Service types generation complete!"

subsystem-test-up:
	$(COMPOSE) -f test/subsystem/docker-compose.yaml up -d --build

subsystem-test-down:
	$(COMPOSE) -f test/subsystem/docker-compose.yaml down -v

subsystem-test:
	go run github.com/onsi/ginkgo/v2/ginkgo -r --randomize-all --fail-on-pending -tags=subsystem ./test/subsystem

subsystem-test-full: subsystem-test-up subsystem-test subsystem-test-down

.PHONY: build run clean fmt vet lint test tidy generate-types generate-spec generate-server generate-client generate-api check-generate-api check-aep generate-service-types subsystem-test-up subsystem-test-down subsystem-test subsystem-test-full
