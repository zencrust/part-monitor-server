FROM golang:alpine AS multistage

RUN apk add --no-cache --update git alpine-sdk

WORKDIR /go/src/alarm-logger
COPY . .

RUN go get -d -v \
  && go install -v \
  && go build -ldflags "-s -w"

FROM alpine:latest
COPY --from=multistage /go/bin/alarm-logger /

EXPOSE 9503
CMD ["./alarm-logger"]