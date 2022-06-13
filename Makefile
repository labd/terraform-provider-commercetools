.PHONY: docs
LOCAL_TEST_VERSION = 99.0.0
OS_ARCH = darwin_arm64

build:
	go build


# Build local provider with very high version number for easier local testing and debugging
# see: https://discuss.hashicorp.com/t/easiest-way-to-use-a-local-custom-provider-with-terraform-0-13/12691/5
build-local:
	go build -o terraform-provider-commercetools_${LOCAL_TEST_VERSION}
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/labd/commercetools/${LOCAL_TEST_VERSION}/${OS_ARCH}
	cp terraform-provider-commercetools_${LOCAL_TEST_VERSION} ~/.terraform.d/plugins/registry.terraform.io/labd/commercetools/${LOCAL_TEST_VERSION}/${OS_ARCH}/terraform-provider-commercetools_v${LOCAL_TEST_VERSION}


format:
	go fmt ./...

test:
	go test -v ./...

update-sdk:
	GO111MODULE=on go get github.com/labd/commercetools-go-sdk
	GO111MODULE=on go mod vendor
	GO111MODULE=on go mod tidy

docs:
	tfplugindocs

coverage-html:
	go test -race -coverprofile=coverage.txt -covermode=atomic -coverpkg=./... ./...
	go tool cover -html=coverage.txt

coverage:
	go test -race -coverprofile=coverage.txt -covermode=atomic -coverpkg=./... ./...
	go tool cover -func=coverage.txt

testacc:
	TF_ACC=1 go test -v ./...

testacct:
	TF_ACC=1 go test -race -coverprofile=coverage.txt -covermode=atomic -coverpkg=./... -v ./...

mockacc:
	TF_ACC=1 \
	CTP_CLIENT_ID=unittest \
	CTP_CLIENT_SECRET=x \
	CTP_PROJECT_KEY=unittest \
	CTP_SCOPES=manage_project:projectkey \
	CTP_API_URL=http://localhost:8989 \
	CTP_AUTH_URL=http://localhost:8989 \
	go test -count=1 -v ./...  
