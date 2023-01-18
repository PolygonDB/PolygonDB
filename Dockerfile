FROM ghcr.io/parkervcp/installers:debian

COPY main.go go.mod go.sum /app/
COPY databases /app/databases
