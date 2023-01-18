FROM golang:latest

COPY main.go go.mod go.sum /app/
COPY databases /app/databases

RUN go build -o main .

CMD ["./main"]
