FROM golang:1.8 as builder
ADD workspace /go
RUN make github-sync

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/bin/linux/server/github-sync /usr/local/bin/github-sync
CMD ["github-sync"]