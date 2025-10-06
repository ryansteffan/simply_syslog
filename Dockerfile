FROM golang:trixie AS builder

WORKDIR /src

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 go build -o /build/simply-syslog /src/cmd/simplysyslog/main.go


FROM scratch AS runner

# # Sets the environment variables used for configuration of the server.
ENV PROTOCOL="UDP"
ENV BIND_ADDRESS="0.0.0.0"
ENV UDP_PORT="514"
ENV TCP_PORT="514"
ENV MAX_TCP_CONNECTIONS="10"
ENV BUFFER_LENGTH="32"
ENV BUFFER_LIFESPAN="5"
ENV MAX_MESSAGE_SIZE="1024"
ENV SYSLOG_PATH="/var/log/syslog.log"
ENV DEBUG_MESSAGES="True"
ENV SYSLOG_FORMATS="[{\"Version\":1,\"Name\":\"RFC3164\",\"Format\":\"^<(?<pri>\\\\d+)>(?<timestamp>\\\\w{3} +\\\\d{1,2} \\\\d{2}:\\\\d{2}:\\\\d{2}) (?<hostname>\\\\S+) (?<tag>\\\\S+?)(?:\\\\[(?<pid>\\\\d+)\\\\])?:? (?<message>.+)$\"},{\"Version\":1,\"Name\":\"RFC5424\",\"Format\":\"^<(?<pri>\\\\d+)>(?<version>\\\\d+) (?<timestamp>[^ ]+) (?<hostname>\\\\S+) (?<appname>\\\\S+) (?<procid>\\\\S+) (?<msgid>\\\\S+) (?<structured_data>(?:\\\\[[^\\\\]]*\\\\]|-)) ?(?<message>.*)$\"}]"

WORKDIR /bin/

COPY --from=builder /build/simply-syslog .

VOLUME [ ${SYSLOG_PATH} ]

EXPOSE ${UDP_PORT}/udp
EXPOSE ${TCP_PORT}/tcp

CMD [ "simply-syslog", "-env", "-env-regex" ]
