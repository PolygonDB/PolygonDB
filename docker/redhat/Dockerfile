FROM        --platform=$TARGETOS/$TARGETARCH redhat/ubi8-micro:latest

LABEL       author="JewishLewish" \
            maintainer="lenny@lunes.host" \
            org.opencontainers.image.source="https://github.com/JewishLewish/PolygonDB" \
            org.opencontainers.image.licenses=MIT

RUN         dnf update -y && dnf install -y iproute gawk sed

USER        container
ENV         USER=container HOME=/home/container

WORKDIR     /home/container

COPY        ./entrypoint.sh /entrypoint.sh
CMD         [ "/bin/bash", "/entrypoint.sh" ]