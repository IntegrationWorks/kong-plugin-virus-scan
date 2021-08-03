FROM kong:2.5.0-ubuntu

COPY /avscanner-client /usr/local/bin/
COPY config.yml /tmp/