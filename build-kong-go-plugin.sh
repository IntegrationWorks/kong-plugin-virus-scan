go mod tidy
go build -o avscanner-client cmd/kong-av-scanner-client/main.go

docker -D build -t kong-go-avscanner-demo .
