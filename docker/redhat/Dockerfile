FROM        --platform=$TARGETOS/$TARGETARCH alpine:3.14

LABEL       author="JewishLewish" \
            maintainer="lenny@lunes.host" \
            org.opencontainers.image.source="https://github.com/JewishLewish/PolygonDB" \
            org.opencontainers.image.licenses=MIT

SHELL       ["/bin/sh", "-eux"]
RUN         apk update --no-cache
RUN         apk add --no-cache libpcap-devel libpcap gcc-c++ openssl-devel \
            openssl-libs openssl ncurses-devel lksctp-tools lksctp-tools-devel

USER        container
ENV         USER=container HOME=/home/container

WORKDIR     /home/container

COPY        ./entrypoint.sh /entrypoint.sh
CMD         [ "/bin/bash", "/entrypoint.sh" ]