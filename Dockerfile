FROM golang:1.8

ADD Makefile /go/Makefile
ADD hack /go/hack
ADD src /go/src
ADD vendor /go/vendor

RUN make build && \
      mv bin/linux/cli/ssh-cli /usr/local/bin/cli && \
      mv bin/linux/server/ssh-server /usr/local/bin/ssh-server && \
      chmod a+x /usr/local/bin/cli && \
      chmod a+x /usr/local/bin/ssh-server

# Cleanup.
RUN rm -fR /go

CMD ["ssh-server"]
