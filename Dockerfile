FROM debian:9

RUN apt-get update && apt-get install -y golang-1.7

COPY . /app/github.com/jemoster/icfp2017
WORKDIR /app
ENV GOPATH=/app
ENV PATH=$PATH:/usr/lib/go-1.7/bin
