##############################################################################
#  Build Xenon
###############################################################################

FROM golang:1.16 as builder

ARG XENON_BRANCH=master
RUN go env -w GO111MODULE=off; 
RUN set -ex; \
    mkdir -p /go/src/github.com/radondb; \
    cd /go/src/github.com/radondb; \
    git clone --branch $XENON_BRANCH https://github.com/radondb/xenon.git; \
    cd xenon;  \
    make build
WORKDIR /workspace
# Copy the go source
COPY cmd/xenon/main.go cmd/xenon/main.go 
COPY utils/ utils/ 
COPY go.mod go.mod
COPY go.sum go.sum
RUN  go env -w GO111MODULE=on && go mod download
# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o xenonchecker cmd/xenon/main.go
###############################################################################
#  Docker image for Xenon
###############################################################################

FROM alpine:3.13

RUN set -ex \
    && addgroup -g 1001 mysql && adduser -u 1001 -g 1001 -S mysql \
    && apk add --no-cache curl bash jq \
    && mkdir -p /etc/xenon /var/lib/xenon /lib64 \
    && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2 \
    && echo "/etc/xenon/xenon.json" > /config.path \
    # allow to change config files
    && chown -R 1001:1001 /etc/xenon /var/lib/xenon

COPY --from=builder /go/src/github.com/radondb/xenon/bin/xenon /usr/local/bin/xenon
COPY --from=builder /go/src/github.com/radondb/xenon/bin/xenoncli /usr/local/bin/xenoncli
COPY --from=builder /workspace/xenonchecker /xenonchecker
USER 1001
WORKDIR /
EXPOSE 8801
VOLUME ["/var/lib/xenon", "/etc/xenon"]

ENTRYPOINT ["xenon"]
CMD ["-c", "/etc/xenon/xenon.json"]