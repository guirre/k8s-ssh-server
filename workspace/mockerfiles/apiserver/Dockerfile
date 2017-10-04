FROM debian:latest 

RUN apt-get update && \
    apt-get install -y curl && \
    apt-get clean

RUN curl -Lo /usr/local/bin/hyperkube https://storage.googleapis.com/kubernetes-release/release/v1.7.0/bin/linux/amd64/hyperkube && chmod +x /usr/local/bin/hyperkube
RUN curl -Lo /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v1.7.0/bin/linux/amd64/kubectl && chmod +x /usr/local/bin/kubectl

CMD ["/usr/local/bin/hyperkube", "apiserver", "--insecure-bind-address", "0.0.0.0", "--service-cluster-ip-range", "242.0.0.0/16", "--runtime-config", "batch/v2alpha1", "--etcd-servers", "http://127.0.0.1:4001", "--storage-backend", "etcd2", "--storage-media-type", "application/json"]
