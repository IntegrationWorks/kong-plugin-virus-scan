go mod tidy
go build -o avscanner-client cmd/kong-plugin-virus-scan/main.go

docker -D build -t kong-go-avscanner-demo .
