FROM alpine:3.14

RUN apk add ca-certificates

COPY bin/remindme /usr/local/bin

ENTRYPOINT ["remindme"]