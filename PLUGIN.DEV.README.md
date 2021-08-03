Exposes a kong instance with the av-plugin installed.

To build (linux):
* Run `build-kong-go-plugin.sh`

To build and run standalone (linux):
* Run `demo.sh`

Notes:
1. The script demo.sh specifies normal boot configuration but also specifies the environment variable `KONG_NGINX_HTTP_CLIENT_BODY_BUFFER_SIZE`. This variable determines the size of the file that can be processed by Nginx/Kong without generating a temp file in disk and returning an error to the plugin. Currently the plugin only processes files that are not written to disk. This variable must be set a value lower than the host/container memory capacity and greater than the size of the largest acceptable file. See [Kong configuration reference](https://docs.konghq.com/gateway-oss/2.5.x/configuration/#nginx_http_client_body_buffer_size) and [advice](https://support.konghq.com/support/s/article/Kong-plugin-produces-a-warning-a-client-request-body-is-buffered-to-a-temporary-file).
2. The script demo.sh specifies the environment variable `KONG_LOG_LEVEL`. Valid values are `debug`, `info`, `notice`, `warn`, `error`, `crit`, `alert`, or `emerg`. The default value in Kong is `notice`. At `info` level, the icap-client libraries debug log events are written to kong's log. See the [kong reference](https://docs.konghq.com/gateway-oss/2.5.x/configuration/#log_level) and [nginx reference](https://nginx.org/en/docs/ngx_core_module.html#error_log) for details.


To monitor:
* `docker exec -it kong-go-avscanner-demo-instance tail -f /usr/local/kong/logs/error.log`
