# build backend
FROM golang:1.19 as go-builder

RUN apt-get update && apt-get install -y git upx

WORKDIR /app
COPY . /app

RUN go generate

ENV CGO_ENABLED=1
ENV GOOS=linux

RUN  VERSION=`git describe --always --tags`; \
     BUILD_TIME=`date -Is`; \
     go build -ldflags="-s -w -X main.buildVersion=$VERSION -X main.buildTime=$BUILD_TIME" -o /go-mqtt-to-influx

RUN upx /go-mqtt-to-influx

# used to generate /etc/passwd /etc/group
RUN groupadd -g 1000 app && \
    useradd -r -u 1000 -g app app
RUN chown app:app /go-mqtt-to-influx

# build final image (use alpine since we need ca-certificate for influx ssl to work as well as glibc)
FROM alpine
USER app
COPY --from=go-builder /go-mqtt-to-influx    /go-mqtt-to-influx
COPY --from=go-builder /etc/group            /etc/group
COPY --from=go-builder /etc/passwd           /etc/passwd
CMD ["/go-mqtt-to-influx", "-c", "/app/config.yaml"]
