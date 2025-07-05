FROM mcr.microsoft.com/devcontainers/go:1-1.23-bookworm
COPY cacert.pem /etc/ssl/certs/cacert.pem
RUN update-ca-certificates
USER root
RUN mkdir -p /home/vscode/go/bin && chown -R vscode:vscode /home/vscode/go
USER vscode
ENV PATH="/home/vscode/go/bin:$PATH"
RUN go install github.com/spf13/cobra-cli@latest