FROM arm32v7/golang:alpine3.9 AS multistage

RUN apk add --no-cache --update alpine-sdk
COPY qemu-arm-static /usr/bin

WORKDIR /go/src/alarm-logger
COPY . .

RUN go get -d -v \
  && go install -v \
  && go build -ldflags "-s -w"

FROM arm32v7/alpine:latest
COPY qemu-arm-static /usr/bin
COPY --from=multistage /go/bin/alarm-logger /
RUN apk add --no-cache tzdata

EXPOSE 9503
CMD ["./alarm-logger"]