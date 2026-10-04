FROM alpine:latest

ARG TARGETPLATFORM

# Install dependencies, create unprivileged user, and set up owned config directory
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S bonjour && \
    adduser -S bonjour -G bonjour -h /home/bonjour && \
    mkdir -p /home/bonjour/.config/bonjour && \
    chown -R bonjour:bonjour /home/bonjour

# Copy binary
COPY $TARGETPLATFORM/bonjour /usr/local/bin/bonjour

# Switch to non-root user
USER bonjour
WORKDIR /home/bonjour

ENTRYPOINT ["/usr/local/bin/bonjour"]