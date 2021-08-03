go mod tidy
go build -o avscanner-client cmd/kong-av-scanner-client/main.go

docker -D build -t kong-go-avscanner-demo .

docker -D run -ti --rm --name kong-go-avscanner-demo-instance \
  -e "KONG_DATABASE=off" \
  -e "KONG_DECLARATIVE_CONFIG=/tmp/config.yml" \
  -e "KONG_PLUGINS=avscanner-client" \
  -e "KONG_PLUGINSERVER_NAMES=avscanner-client" \
  -e "KONG_PLUGINSERVER_AVSCANNER_CLIENT_QUERY_CMD=avscanner-client -dump" \
  -e "KONG_PROXY_LISTEN=0.0.0.0:8000" \
  -e "KONG_NGINX_HTTP_CLIENT_BODY_BUFFER_SIZE=100k" \
  -e "KONG_LOG_LEVEL=debug" \
  -p 8000:8000 \
  -p 8001:8001 \
  kong-go-avscanner-demo