FROM --platform=$TARGETOS/$TARGETARCH alpine:latest

RUN apk update && apk add git go

ENV GOPATH /go
ENV PATH $GOPATH/bin:$PATH

RUN git clone https://github.com/JewishLewish/PolygonDB /tmp/tmp
RUN mv /tmp/tmp $GOPATH/home/container

WORKDIR $GOPATH/home/container
RUN go build .