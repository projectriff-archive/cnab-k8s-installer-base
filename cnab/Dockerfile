FROM ubuntu:18.04 as kapp
RUN apt-get update && apt-get install -y curl
RUN curl -Ls https://github.com/k14s/kapp/releases/download/v0.9.0/kapp-linux-amd64 > /usr/local/bin/kapp \
  && chmod 755 /usr/local/bin/kapp

FROM ubuntu:18.04
COPY app/run /cnab/app/run
COPY --from=kapp /usr/local/bin/kapp /usr/local/bin/kapp
CMD ["/cnab/app/run"]
