FROM golang:alpine AS build

ARG SOURCE_COMMIT

ADD . /go/src/github.com/bobtfish/aws-instance-health-exporter
WORKDIR /go/src/github.com/bobtfish/aws-instance-health-exporter

RUN DATE=$(date -u '+%Y-%m-%d-%H%M UTC'); \
    go install -ldflags="-X 'main.Version=${SOURCE_COMMIT}' -X 'main.BuildTime=${DATE}'" ./...

FROM alpine:latest
COPY --from=build /go/bin/aws-instance-health-exporter /bin

ENTRYPOINT  [ "/bin/aws-instance-health-exporter" ]
EXPOSE      9383
