FROM golang:1.8 as builder
RUN go get github.com/mitchellh/gox
ADD workspace /go
RUN make

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/bin/k8s-ssh-server-server_linux_amd64 /usr/local/bin/ssh-server
COPY --from=builder /go/bin/k8s-ssh-server-cli_linux_amd64 /usr/local/bin/ssh-cli
CMD ["ssh-server"]
