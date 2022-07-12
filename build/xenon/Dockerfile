FROM golang:1.16 as builder

WORKDIR /
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go env -w GOPROXY=https://goproxy.cn,direct; 
#     go mod download
RUN  go mod download

# Copy the go source
COPY cmd/xenon/main.go cmd/xenon/main.go 
COPY utils/ utils/ 

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o bin/xenonchecker cmd/xenon/main.go

FROM radondb/xenon:1.1.5-alpha

COPY --from=builder /bin/xenonchecker /xenonchecker
