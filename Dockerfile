FROM golang:trixie AS builder

WORKDIR /src

COPY . .

RUN go mod download

RUN go build -o /build/simply-syslog /src/cmd/simplysyslog/main.go


FROM ubuntu:latest AS runner

# # Sets the enviroment variables used for configuration of the server.
ENV PROTOCOL="UDP"
ENV BIND_ADDRESS="0.0.0.0"
ENV UDP_PORT="514"
ENV TCP_PORT="514"
ENV MAX_TCP_CONNECTIONS="10"
ENV BUFFER_LENGTH="32"
ENV BUFFER_LIFESPAN="5"
ENV MAX_MESSAGE_SIZE="1024"
ENV SYSLOG_PATH="/var/log/simply_syslog/"
ENV DEBUG_MESSAGES="True"

WORKDIR /bin/

COPY --from=builder /build/simply-syslog .

CMD [ "simply-syslog", "-env" ]
