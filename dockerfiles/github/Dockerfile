FROM golang:1.8 as builder
RUN go get github.com/mitchellh/gox
ADD workspace /go
RUN make github

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/bin/github_linux_amd64 /usr/local/bin/github
CMD ["github"]
