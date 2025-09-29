FROM golang:trixie as Builder

WORKDIR /src

COPY . .

RUN go mod download

RUN go build -o /build/simply-syslog /src/cmd/simplysyslog/main.go


FROM scratch as Runner

# # Sets the enviroment variables used for configuration of the server.
ENV PROTOCOL "UDP"
ENV BIND_ADDRESS "0.0.0.0"
ENV UDP_PORT "514"
ENV TCP_PORT "514"
ENV MAX_TCP_CONNECTIONS "10"
ENV BUFFER_LENGTH "32"
ENV BUFFER_LIFESPAN "5"
ENV MAX_MESSAGE_SIZE "1024"
ENV SYSLOG_PATH "/var/log/simply_syslog/"
ENV DEBUG_MESSAGES "True"

WORKDIR /bin/

COPY --from=Builder /build/simply-syslog .

CMD [ "simply-syslog" ]

# FROM python

# # Sets the enviroment variables used for configuration of the server.
# ENV PROTOCOL "UDP"
# ENV BIND_ADDRESS "0.0.0.0"
# ENV UDP_PORT "514"
# ENV TCP_PORT "514"
# ENV MAX_TCP_CONNECTIONS "10"
# ENV BUFFER_LENGTH "32"
# ENV BUFFER_LIFESPAN "5"
# ENV MAX_MESSAGE_SIZE "1024"
# ENV SYSLOG_PATH "/var/log/simply_syslog/"
# ENV DEBUG_MESSAGES "True"

# # Makes the needed paths for storing the server and storing the log file.
# RUN mkdir "/simply_syslog/"
# RUN mkdir $SYSLOG_PATH

# # Copys the code for the server to the docker image.
# COPY config/server_logging_config.json "/simply_syslog/config/"
# COPY src "/simply_syslog/src/"
# COPY init.py "/simply_syslog/"
# COPY main.py "/simply_syslog/"

# # Makes a volume used to save the file that stores the syslogs from remote hosts.
# VOLUME $SYSLOG_PATH

# # Exposes the needed ports for the server.
# EXPOSE $UDP_PORT/udp
# EXPOSE $TCP_PORT/tcp

# # Sets a work dir for where the CMD is to run from.
# WORKDIR "/simply_syslog/"

# # Runs the init.py file to set up the server config, and start the server.
# CMD ["python3", "./init.py"]