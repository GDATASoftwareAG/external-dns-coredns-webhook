FROM golang:1.19-buster as builder

COPY . /code/external-dns-coredns-plugin
WORKDIR /code/external-dns-coredns-plugin
RUN CGO_ENABLED=0 go build

FROM debian:buster-slim

COPY --from=builder /code/external-dns-coredns-plugin/external-dns-coredns-plugin /usr/bin/external-dns-coredns-plugin

# replace with your desire device count
CMD ["external-dns-coredns-plugin"]

LABEL org.opencontainers.image.title="ExternalDNS CoreDNS plugin Docker Image" \
      org.opencontainers.image.description="external-dns-coredns-plugin" \
      org.opencontainers.image.url="https://github.com/GDATASoftwareAG/external-dns-coredns-plugin" \
      org.opencontainers.image.source="https://github.com/GDATASoftwareAG/external-dns-coredns-plugin" \
      org.opencontainers.image.license="MIT"