FROM mcr.microsoft.com/devcontainers/go:1-1.23-bookworm
COPY cacert.pem /etc/ssl/certs/cacert.pem
RUN update-ca-certificates
RUN go install github.com/spf13/cobra-cli@latest