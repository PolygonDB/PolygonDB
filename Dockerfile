FROM golang:latest

COPY main.go go.mod go.sum ./
COPY databases/ ./databases/
