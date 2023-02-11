FROM        --platform=$TARGETOS/$TARGETARCH alpine:3.14

LABEL       author="JewishLewish" \
            maintainer="lenny@lunes.host" \
            org.opencontainers.image.source="https://github.com/JewishLewish/PolygonDB" \
            org.opencontainers.image.licenses=MIT


RUN         apk add --update --no-cache ca-certificates tzdata gcc g++ libgcc1 libc++-dev gdb libc6 git wget curl tar zip unzip\
				&& adduser -D -h /home/container container

USER        container
ENV         USER=container HOME=/home/container

WORKDIR     /home/container

COPY        ./entrypoint.sh /entrypoint.sh
CMD         [ "/bin/bash", "/entrypoint.sh" ]