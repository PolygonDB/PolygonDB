FROM ghcr.io/parkervcp/yolks:debian as base

COPY main.go go.mod go.sum /app/
COPY database /app/database
