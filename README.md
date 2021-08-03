Kong Plugin - Virus Scan
========================

## Summary
- [Kong Plugin - Virus Scan](#kong-plugin---virus-scan)
  - [Summary](#summary)
  - [- HTTP Header Limitation](#--http-header-limitation)
- [Functional Overview](#functional-overview)
- [Configuration Reference](#configuration-reference)
  - [Enable the plugin on a service](#enable-the-plugin-on-a-service)
  - [Enable the plugin globally](#enable-the-plugin-globally)
  - [Parameters](#parameters)
- [Usage](#usage)
  - [Body Types](#body-types)
  - [Logging](#logging)
- [Known limitations](#known-limitations)
  - [File Sizes](#file-sizes)
  - [Limited Testing](#limited-testing)
  - [WIP Client Libraries](#wip-client-libraries)
  - [HTTP Header Limitation](#http-header-limitation)
---

# Functional Overview

Scan file attachments or HTTP request bodies by integrating with an ICAP enabled antivirus server.

> ***Note:*** Although the transport layer of ICAP is standardised in [RFC 3507](https://datatracker.ietf.org/doc/html/rfc3507), the exact format of response messages and headers is not. Athough _most_ commercially available antivirus scanners follow certain conventions - which this plugin has been designed to recognise - it is not guaranteed to be compatible with all, and has been tested solely with [ClamAV](https://www.clamav.net/).

---

# Configuration Reference

This plugin is compatible with requests with the following protocols:

* http
* https

This plugin is compatible with DB-less mode.

## Enable the plugin on a service

For example, configure this plugin on a [service](https://docs.konghq.com/gateway-oss/2.5.x/admin-api/#service-object) by making the following request:

```bash
curl -X POST http://{HOST}:8001/services/{SERVICE}/plugins \
    --data "name=virus-scan"  \
    --data "config.ScannerURL=icap://{AV_HOST}:{ICAP_PORT}/avscan" \
```

## Enable the plugin globally

A plugin which is not associated to any service, route, or consumer is considered *global*, and will be run on every request. Read the [Plugin Reference](https://docs.konghq.com/gateway-oss/2.5.x/admin-api/#add-plugin) and the [Plugin Precedence](https://docs.konghq.com/gateway-oss/2.5.x/admin-api/#precedence) sections for more information.

For example, configure this plugin globally with:

```bash
$ curl -X POST http://{HOST}:8001/plugins/ \
    --data "name=virus-scan"  \
    --data "config.ScannerURL=icap://{AV_HOST}:{ICAP_PORT}/avscan" \
```

## Parameters

Here's a list of all the parameters which can be used in this plugin's configuration:

| Parameter | Type | Default | Description | Example |
| --------- | ---- | ------- | ----------- | ------- |
| ScannerURL | `String` | none | The ICAP URL that the antivirus server is accessible on. The protocol, host, and port sections of the URL are mandatory. | `icap://icap.example.org:1433/avscan` |

# Usage

When enabled, the plugin will automatically forward any HTTP request bodies to the configured `ICAP` enabled anti-virus server for scanning. Should connectivity to this server fail, or the response from the server indicates an infected payload, the request will be terminated with an `HTTP 400` returned to the API client. 

A request which has been sent successfully to, and been cleared by, the anti-virus server will be forwarded to the upstream service. The response from the upstream service will be proxied back to the API client.

## Body Types

* If the `Content-Type` header identifies multiple files (`multipart`) then each part is sent to the virus scanner as a separate file. All parts must report no infection for processing to continue.
* For non-multipart bodies, the HTTP payload is sent as a single file to the virus scanner.
* The plugin will forward all HTTP payloads, regardless of the `Content-Type` header to the configured virus scanner.

## Logging

The plugin employees the Kong logging subsystem. The Kong configuration parameter `log_level` and its corresponding environment variable `KONG_LOG_LEVEL` determine the level of logs which are written to the Kong log files. Valid values are `debug`, `info`, `notice`, `warn`, `error`, `crit`, `alert`, or `emerg`. The default value in Kong is `notice`. At `info` level, the icap-client libraries debug log events are written to kong's log. See the [kong reference](https://docs.konghq.com/gateway-oss/2.5.x/configuration/#log_level) and [nginx reference](https://nginx.org/en/docs/ngx_core_module.html#error_log) for details.

# Known limitations

## File Sizes

The Kong configuration parameter `nginx_http_client_body_buffer_size` and its corresponding environment variable `KONG_NGINX_HTTP_CLIENT_BODY_BUFFER_SIZE` determines the size of a request that can be processed in memory by Kong/NGINX before generating a temorary file on disk. 

This variable must be set a value lower than the host/container memory capacity and greater than the size of the largest acceptable file. See [Kong configuration reference](https://docs.konghq.com/gateway-oss/2.5.x/configuration/#nginx_http_client_body_buffer_size) and [advice](https://support.konghq.com/support/s/article/Kong-plugin-produces-a-warning-a-client-request-body-is-buffered-to-a-temporary-file).

The plugin also does not implement file chunking - so when a large file is processed, the whole file is loaded in the plugin memory and sent in one small block followed by the rest of the file to the AV scanner.

## Limited Testing

This plugin has been tested using [ClamAV](https://www.clamav.net/) as the antivirus server and the [EICAR test file](https://en.wikipedia.org/wiki/EICAR_test_file) as a safe way of generating positive virus scan results. No genuine viruses have been tested during the implementation of this plugin.

## WIP Client Libraries

This plugin uses the [Egirna ICAP Client](https://github.com/egirna/icap-client) library for integration with the antivirus server. This library is under active development.

## HTTP Header Limitation

This plugin does not support HTTP requests with more than 100 headers sent by the API client.


