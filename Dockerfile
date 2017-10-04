FROM golang:1.8
LABEL maintainer="otiai10 <otiai10@gmail.com>"

COPY . $GOPATH/src/github.com/seqpod/seqpod-api
WORKDIR $GOPATH/src/github.com/seqpod/seqpod-api

RUN go get .
RUN go install .

ENTRYPOINT seqpod-api
