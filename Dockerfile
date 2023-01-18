FROM        --platform=$TARGETOS/$TARGETARCH debian:bullseye-slim

RUN         useradd -m -d /home/container -s /bin/bash container

USER        container
ENV         USER=container HOME=/home/container
ENV         DEBIAN_FRONTEND noninteractive

WORKDIR     /home/container

CMD         [ "cd", "/home/container" ]