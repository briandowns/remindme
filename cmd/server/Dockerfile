FROM alpine:3.14

RUN apk add ca-certificates

COPY bin/remindmed /usr/local/bin

ENTRYPOINT ["remindmed", "&"]
