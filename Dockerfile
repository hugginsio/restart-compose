FROM alpine:3.22.1@sha256:4bcff63911fcb4448bd4fdacec207030997caf25e9bea4045fa6c8c44de311d1

LABEL org.opencontainers.image.title="restart-compose"
LABEL org.opencontainers.image.description="Docker Compose restart utility"
LABEL org.opencontainers.image.authors="hugginsio"
LABEL org.opencontainers.image.url="https://github.com/hugginsio/restart-compose"
LABEL org.opencontainers.image.source="https://github.com/hugginsio/restart-compose"
LABEL org.opencontainers.image.vendor="hugginsio"
LABEL org.opencontainers.image.licenses="BSD-3-Clause"

# Install system dependencies
RUN apk add --no-cache \
    git \
    docker-cli \
    docker-compose

# Set working directory
WORKDIR /app

ADD dist/restart-compose_linux_amd64_v1/restart-compose /app/restart-compose
RUN chmod +x /app/restart-compose

ENTRYPOINT ["/app/restart-compose", "-d", "/data"]
