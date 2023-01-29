FROM --platform=$TARGETOS/$TARGETARCH alpine:latest

RUN apk update && apk add --no-cache go

ENV GOPATH /go
ENV PATH $GOPATH/bin:$PATH

WORKDIR $GOPATH

COPY . .

RUN go build -o /home/container .

CMD ["/home/container"]