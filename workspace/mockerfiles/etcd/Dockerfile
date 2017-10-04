FROM alpine:latest

RUN apk --no-cache add ca-certificates

ADD https://s3-ap-southeast-2.amazonaws.com/bins.skpr.io/server/etcd-v2.2.2 /usr/local/bin/etcd
RUN chmod a+x /usr/local/bin/etcd

CMD ["/usr/local/bin/etcd", "--force-new-cluster"]
