build:
	go build

format:
	go fmt ./...

test:
	go test -v ./...

coverage-html:
	go test -race -coverprofile=coverage.txt -covermode=atomic -coverpkg=./... ./...
	go tool cover -html=coverage.txt

coverage:
	go test -race -coverprofile=coverage.txt -covermode=atomic -coverpkg=./... ./...
	go tool cover -func=coverage.txt

testacc:
	TF_ACC=1 go test -v ./...

mockacc:
	TF_ACC=1 CTP_AUTH_URL=http://localhost:8989 CTP_PROJECT_KEY=unittest CTP_API_URL=http://localhost:8989 CTP_CLIENT_ID=unittest CTP_CLIENT_SECRET=x go test -count=1 -v ./...
