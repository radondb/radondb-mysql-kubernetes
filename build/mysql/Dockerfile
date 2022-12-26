# Build the manager binary
ARG MYSQL_IMAGE
FROM golang:1.17.13 as builder
WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
ARG GO_PROXY=off
RUN if [ "$GO_PROXY" = "on" ]; then \
    go env -w GOPROXY=https://goproxy.cn,direct; \
    fi
RUN  go mod download
# Copy the go source
COPY ./ .
# Build
RUN CGO_ENABLED=0 go build  -ldflags="-s -w" -a -o mysqlchecker ./cmd/mysql/main.go
FROM percona/percona-server:${MYSQL_IMAGE}
WORKDIR /
COPY --from=builder /workspace/mysqlchecker /usr/bin/mysqlchecker
USER 65532:65532
ENTRYPOINT ["/docker-entrypoint.sh"]

USER mysql
EXPOSE 3306
CMD ["mysqld"]
