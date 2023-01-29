FROM --platform=$TARGETOS/$TARGETARCH alpine:latest


RUN apk update && apk add git go gcc

RUN apk add --no-cache --virtual .build-deps curl binutils build-base gcc abuild binutils musl-dev glibc-dev


ENV GOPATH /go
ENV PATH $GOPATH/bin:$PATH

RUN git clone https://github.com/JewishLewish/PolygonDB.git /tmp/PolygonDB
RUN mv /tmp/PolygonDB /home/container

WORKDIR /home/container

RUN go build -o /go/bin/PolygonDB .

CMD ["/go/bin/PolygonDB"]