FROM busybox:1.32.0-glibc

COPY powerwall-exporter /usr/bin/

ENTRYPOINT ["/usr/bin/powerwall-exporter"]

