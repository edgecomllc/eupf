FROM alpine:3.17

WORKDIR /app

RUN apk add --no-cache iptables

COPY ./busy-poll.sh /app/bin/busy-poll.sh

# CMD is overridden if arguments are passed.
ENTRYPOINT [ "sh", "/app/bin/busy-poll.sh" ]