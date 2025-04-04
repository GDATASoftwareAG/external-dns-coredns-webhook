FROM golang:1.24-bookworm AS builder

COPY . /code/external-dns-coredns-webhook
WORKDIR /code/external-dns-coredns-webhook
RUN CGO_ENABLED=0 go build

FROM debian:bookworm-slim

COPY --from=builder /code/external-dns-coredns-webhook/external-dns-coredns-webhook /usr/bin/external-dns-coredns-webhook

# replace with your desire device count
CMD ["external-dns-coredns-webhook"]

LABEL org.opencontainers.image.title="ExternalDNS CoreDNS webhook Docker Image" \
      org.opencontainers.image.description="external-dns-coredns-webhook" \
      org.opencontainers.image.url="https://github.com/GDATASoftwareAG/external-dns-coredns-webhook" \
      org.opencontainers.image.source="https://github.com/GDATASoftwareAG/external-dns-coredns-webhook" \
      org.opencontainers.image.license="MIT"