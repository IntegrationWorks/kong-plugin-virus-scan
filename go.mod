module github.com/IntegrationWorks/kong-plugin-virus-scan

go 1.16

replace github.com/integrationworks/icap-client v0.1.2 => ./icap-client

require (
	github.com/Kong/go-pdk v0.6.1
	github.com/integrationworks/icap-client v0.1.2
)
