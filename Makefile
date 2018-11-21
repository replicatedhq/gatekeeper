
# Image URL to use all building/pushing image targets
IMG ?= controller:latest

all: test manager cli

# Build the CLI
cli: fmt vet test generate
	go build -o bin/gatekeeper github.com/replicatedhq/gatekeeper/cmd/gatekeeper

# Run tests
test: generate fmt vet manifests
	go test ./pkg/... ./cmd/... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager github.com/replicatedhq/gatekeeper/cmd/manager

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet
	go run ./cmd/manager/main.go

# Install CRDs into a cluster
install: manifests
	kubectl apply -f config/crds

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	kubectl apply -f config/crds
	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests:
	go run vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go all

# Run go fmt against code
fmt:
	go fmt ./pkg/... ./cmd/...

# Run go vet against code
vet:
	go vet ./pkg/... ./cmd/...

# Generate code
generate:
	go generate ./pkg/... ./cmd/...
	rm -rf ./pkg/client/gatekeeperclientset/fake
	rm -rf ./pkg/client/gatekeeperclientset/typed/policies/v1alpha2/fake

# Build the docker image
docker-build: test
	docker build . -t ${IMG}
	@echo "updating kustomize image patch file for manager resource"
	sed -i'' -e 's@image: .*@image: '"${IMG}"'@' ./config/default/manager_image_patch.yaml

# Push the docker image
docker-push:
	docker push ${IMG}

.state/vet: $(SRC)
	go vet ./pkg/...
	go vet ./cmd/...
	@mkdir -p .state
	@touch .state/vet

.state/cc-test-reporter:
	@mkdir -p .state/
	wget -O .state/cc-test-reporter https://codeclimate.com/downloads/test-reporter/test-reporter-latest-linux-amd64
	chmod +x .state/cc-test-reporter

.state/lint: $(SRC)
	golint ./pkg/... | grep -vE '_mock|e2e' | grep -v "should have comment" | grep -v "comment on exported" || :
	golint ./cmd/... | grep -vE '_mock|e2e' | grep -v "should have comment" | grep -v "comment on exported" || :
	@mkdir -p .state
	@touch .state/lint

.state/coverage.out: $(SRC)
	@mkdir -p .state/
	#the reduced parallelism here is to avoid hitting the memory limits - we consistently did so with two threads on a 4gb instance
	go test -parallel 1 -p 1 -coverprofile=.state/coverage.out ./pkg/...

ci-upload-coverage: .state/coverage.out .state/cc-test-reporter
	./.state/cc-test-reporter format-coverage -o .state/codeclimate/codeclimate.json -t gocov .state/coverage.out
	./.state/cc-test-reporter upload-coverage -i .state/codeclimate/codeclimate.json

citest: .state/vet .state/lint .state/coverage.out
