FROM scratch

ADD gcp-status-exporter /gcp-status-exporter

ENTRYPOINT ["/gcp-status-exporter"]
