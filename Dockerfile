FROM golang:1.8
LABEL maintainer="otiai10 <otiai10@gmail.com>"

COPY . $GOPATH/src/github.com/otiai10/fastpot-api
WORKDIR $GOPATH/src/github.com/otiai10/fastpot-api

RUN go get .
RUN go install .

ENTRYPOINT fastpot-api
