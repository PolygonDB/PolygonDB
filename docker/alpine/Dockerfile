FROM        --platform=$TARGETOS/$TARGETARCH golang:1.19.5 as build

RUN adduser -D -h /home/container container 

USER container
ENV  HOME /home/container
WORKDIR /home/container

COPY . .
RUN go build


FROM        --platform=$TARGETOS/$TARGETARCH alpine:latest

USER container
ENV  HOME /home/container
WORKDIR /home/container

CMD ["./PolygonDB"]