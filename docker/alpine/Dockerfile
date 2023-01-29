FROM --platform=$TARGETOS/$TARGETARCH golang:1.15-alpine3.12


RUN apk update && apk add git go gcc


ENV GOPATH /go
ENV PATH $GOPATH/bin:$PATH

RUN git clone https://github.com/JewishLewish/PolygonDB.git /tmp/PolygonDB
RUN mv /tmp/PolygonDB /home/container

WORKDIR /home/container

RUN go build -o /go/bin/PolygonDB .

CMD ["/go/bin/PolygonDB"]