FROM golang:1.8 as builder
ADD workspace /go
RUN make server
RUN make cli

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/bin/linux/server/ssh-server /usr/local/bin/ssh-server
COPY --from=builder /go/bin/linux/cli/ssh-cli /usr/local/bin/ssh-cli
CMD ["ssh-server"]