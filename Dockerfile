FROM debian:9

RUN apt-get update && apt-get install -y golang-1.7 git python3

RUN mkdir -p /src
WORKDIR /src
ENV GOPATH=/
ENV PATH=$PATH:/usr/lib/go-1.7/bin
RUN go get \
    github.com/golang/glog \
    gonum.org/v1/gonum/graph \
    gonum.org/v1/gonum/graph/internal/set \
    gonum.org/v1/gonum/graph/internal/ordered \
    gonum.org/v1/gonum/blas \
    gonum.org/v1/gonum/internal/asm/c128 \
    gonum.org/v1/gonum/internal/asm/f32 \
    gonum.org/v1/gonum/floats \
    gonum.org/v1/gonum/blas/gonum \
    gonum.org/v1/gonum/lapack \
    gonum.org/v1/gonum/blas/blas64 \
    gonum.org/v1/gonum/lapack/gonum \
    gonum.org/v1/gonum/lapack/lapack64 \
    gonum.org/v1/gonum/mat \
    gonum.org/v1/gonum/graph/simple \
    gonum.org/v1/gonum/graph/internal/linear \
    gonum.org/v1/gonum/graph/traverse \
    gonum.org/v1/gonum/graph/path

ENV PYTHONUNBUFFERED=1

# Everything after this line will get run every time!
COPY . /src/github.com/jemoster/icfp2017
WORKDIR /src/github.com/jemoster/icfp2017
RUN go get -v github.com/jemoster/icfp2017/...
RUN go build github.com/jemoster/icfp2017/src/bots/unremarkable/simpleton
RUN go build github.com/jemoster/icfp2017/src/bots/prattmic/walk
RUN go build github.com/jemoster/icfp2017/src/bots/akesling/brownian
RUN go build github.com/jemoster/icfp2017/src/bots/akesling/strategery
RUN go build github.com/jemoster/icfp2017/src/bots/cdfox/blob
RUN go build github.com/jemoster/icfp2017/src/bots/punter76
RUN go build github.com/jemoster/icfp2017/src/bots/punter2
