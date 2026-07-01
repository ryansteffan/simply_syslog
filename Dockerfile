FROM golang:1.25.1-bookworm AS builder

WORKDIR /src

COPY . .

RUN apt-get update && apt-get install -y --no-install-recommends build-essential

RUN go mod download

RUN CGO_ENABLED=1 go build -o /build/simply-syslog /src/cmd/simplysyslog/main.go


FROM debian:bookworm-slim AS runner

# UDP Server Config
ENV UDP_SERVER_ENABLED="true"
ENV UDP_BIND_ADDRESS="0.0.0.0"
ENV UDP_PORT="514"
ENV UDP_MAX_MESSAGE_SIZE="1024"

# TCP Server Config
ENV TCP_SERVER_ENABLED="false"
ENV TCP_BIND_ADDRESS="0.0.0.0"
ENV TCP_PORT="514"
ENV TCP_MAX_MESSAGE_SIZE="1024"

# TLS Server Config
ENV TLS_SERVER_ENABLED="false"
ENV TLS_BIND_ADDRESS="0.0.0.0"
ENV TLS_PORT="6514"
ENV TLS_MAX_MESSAGE_SIZE="1024"

# Logging and buffer settings
ENV SELF_LOGGING_LEVEL="7"
ENV BUFFER_MAX_ITEMS="1024"
ENV BUFFER_MAX_LIFETIME="15"

WORKDIR /simply_syslog/

COPY --from=builder /build/simply-syslog .

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates libstdc++6 && rm -rf /var/lib/apt/lists/*

# Map volumes for configuration, logs, and syslog database data.
VOLUME [ "/simply_syslog/config", "/simply_syslog/logs", "/syslog/" ]

EXPOSE ${UDP_PORT}/udp
EXPOSE ${TCP_PORT}/tcp
EXPOSE ${TLS_PORT}/tcp

CMD [ "./simply-syslog", "-env-gen-config" ]
