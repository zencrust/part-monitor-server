FROM arm32v7/golang:alpine3.9 AS multistage

RUN apk add --no-cache --update alpine-sdk

WORKDIR /go/src/alarm-logger
COPY . .

RUN go get -d -v \
  && go install -v \
  && go build -ldflags "-s -w"

FROM arm32v7/alpine:latest
COPY --from=multistage /go/bin/alarm-logger /

EXPOSE 9503
CMD ["./alarm-logger"]