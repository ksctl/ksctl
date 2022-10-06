FROM golang:1.19

WORKDIR /go/src

COPY go.sum .

COPY go.mod .

RUN go get -d || true

COPY . .

RUN go build -o kubesimpctl

RUN cp -v kubesimpctl /usr/local/bin

RUN mkdir /etc/kubesimpctl

RUN mkdir -p /home/$(whoami)/.kube/kubesimpctl/config && \
    mkdir -p /home/$(whoami)/.kube/kubesimpctl/cred && \
    touch /home/$(whoami)/.kube/kubesimpctl/cred/aws && \
    touch /home/$(whoami)/.kube/kubesimpctl/cred/azure && \
    touch /home/$(whoami)/.kube/kubesimpctl/cred/civo && \
    mkdir /home/$(whoami)/.kube/kubesimpctl/config/azure && \
    mkdir /home/$(whoami)/.kube/kubesimpctl/config/aws && \
    mkdir /home/$(whoami)/.kube/kubesimpctl/config/local
