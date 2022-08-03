FROM golang:1.16 as builder

WORKDIR /
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN if [ $(date +%z) = "+0800" ] ; then go env -w GOPROXY=https://goproxy.cn,direct; fi
# RUN go env -w GOPROXY=https://goproxy.cn,direct
#     go mod download
RUN  go mod download

# Copy the go source
COPY cmd/mysql/main.go main.go 
COPY internal/ internal/
COPY api/ api/
COPY utils/ /utils/
COPY mysqlcluster/ mysqlcluster/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o bin/mysqlchecker main.go

FROM percona/percona-server:5.7.34

COPY --from=builder /bin/mysqlchecker /mysqlchecker