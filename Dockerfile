FROM ghcr.io/parkervcp/installers:debian as base

COPY main.go go.mod go.sum /app/
COPY databases /app/databases
