FROM        --platform=$TARGETOS/$TARGETARCH alpine:latest

LABEL       author="JewishLewish" \
            maintainer="lenny@lunes.host" \
            org.opencontainers.image.source="https://github.com/JewishLewish/PolygonDB" \
            org.opencontainers.image.licenses=MIT


RUN         apk add --update --no-cache bash ca-certificates tzdata curl git jq wget libc6-compat\
				&& adduser -D -h /home/container container

USER        container
ENV         USER=container HOME=/home/container

WORKDIR     /home/container

COPY        ./entrypoint.sh /entrypoint.sh
CMD         [ "/bin/bash", "/entrypoint.sh", "-l", "-c" ]