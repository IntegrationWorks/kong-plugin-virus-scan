FROM kong:2.8.1-ubuntu

COPY /avscanner-client /usr/local/bin/
COPY config.yml /tmp/