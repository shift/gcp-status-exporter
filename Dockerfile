FROM scratch

ADD gcp-status-exporter /gcp-status-exporter
ADD ca-bundle.crt /etc/ssl/certs/ca-bundle.crt
ENTRYPOINT ["/gcp-status-exporter"]
