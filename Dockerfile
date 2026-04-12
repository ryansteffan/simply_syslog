FROM golang:1.25.1-bookworm AS builder

WORKDIR /src

COPY . .

RUN apt-get update && apt-get install -y --no-install-recommends build-essential

RUN go mod download

RUN CGO_ENABLED=1 go build -o /build/simply-syslog /src/cmd/simplysyslog/main.go


FROM debian:bookworm-slim AS runner

# Sets the environment variables used for configuration of the server.
ENV SERVER_MODE="udp"
ENV BIND_ADDRESS="0.0.0.0"
ENV UDP_PORT="514"
ENV TCP_PORT="514"
ENV TLS_PORT="6514"
ENV SELF_LOGGING_LEVEL=6 

WORKDIR /simply_syslog/

COPY --from=builder /build/simply-syslog .

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates libstdc++6 && rm -rf /var/lib/apt/lists/*

# Map volumes for configuration, logs, and syslog database data.
VOLUME [ "/simply_syslog/config", "/simply_syslog/logs", "/syslog/" ]

EXPOSE ${UDP_PORT}/udp
EXPOSE ${TCP_PORT}/tcp
EXPOSE ${TLS_PORT}/tcp

CMD [ "./simply-syslog", "-env-gen-config" ]
