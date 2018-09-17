build:
	go build

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