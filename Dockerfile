FROM ghcr.io/parkervcp/yolks:debian

COPY main.go go.mod go.sum /app/
COPY databases /app/databases
