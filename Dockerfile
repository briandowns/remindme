FROM golang:1.17.2 as BUILDER

WORKDIR /build
ADD . /build/
RUN go mod tidy && make

FROM alpine:3.14

RUN apk add ca-certificates

COPY --from=BUILDER /build/bin/remindme /usr/local/bin

ENTRYPOINT ["remindme"]
