#
# Simple tool to watch directory for new files and upload them to S3
#

FROM golang:1.24.2 AS test
WORKDIR /build
ENV GOPATH=/go
ENV PATH="$PATH:$GOPATH/bin"
COPY Makefile Makefile
COPY *.go ./
COPY go.mod go.mod
COPY go.sum go.sum
COPY internal/ internal/
COPY fixtures/ fixtures/
RUN make test

FROM test AS build
WORKDIR /build
ENV GOPATH=/go
ENV PATH="$PATH:$GOPATH/bin"
RUN make build

# FROM gcr.io/distroless/base-debian11
FROM alpine:3.21
WORKDIR /
COPY --from=build /build/output/pd-cert-assistant /pd-cert-assistant
ENTRYPOINT ["/pd-cert-assistant"]
