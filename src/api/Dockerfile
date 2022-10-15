FROM golang:1.19 AS civoTest

WORKDIR /go/src

COPY go.sum .

COPY go.mod .

RUN go get -d || true

COPY . .

RUN mkdir -p /home/$(whoami)/.kube/kubesimpctl/config/civo && \
mkdir -p /home/$(whoami)/.kube/kubesimpctl/cred && \
touch /home/$(whoami)/.kube/kubesimpctl/cred/civo


WORKDIR /go/src/civo

CMD ["go", "test", "-v", "."]