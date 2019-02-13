FROM golang:1.11.5-alpine

RUN apk --no-cache add curl \
  && curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl \
  && mv ./kubectl /usr/local/bin \
  && chmod 755 /usr/local/bin/kubectl \
  && apk del curl

COPY Dockerfile /cnab/Dockerfile

CMD ["/cnab/app/run"]
