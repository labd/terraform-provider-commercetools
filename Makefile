build:
	go build

format:
	go fmt ./...

test:
	go test -v ./...

update-sdk:
	GO111MODULE=on go get github.com/labd/commercetools-go-sdk
	GO111MODULE=on go mod vendor
	GO111MODULE=on go mod tidy

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
	CTP_CLIENT_ID=clientid \
	CTP_CLIENT_SECRET=clientsecret \
	CTP_PROJECT_KEY=projectkey \
	CTP_SCOPES=view_project:projectkey \
	CTP_API_URL=http://localhost:8989 \
	CTP_AUTH_URL=http://localhost:8989 \
	go test -count=1 -v ./...
