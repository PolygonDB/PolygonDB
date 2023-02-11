FROM        --platform=$TARGETOS/$TARGETARCH centos:centos7

LABEL       author="JewishLewish" \
            maintainer="lenny@lunes.host" \
            org.opencontainers.image.source="https://github.com/JewishLewish/PolygonDB" \
            org.opencontainers.image.licenses=MIT

RUN         yum -y update
RUN         yum -y install libpcap-devel libpcap gcc-c++ openssl-devel \
            openssl-libs openssl ncurses-devel lksctp-tools lksctp-tools-devel

RUN         yum clean all

USER        container
ENV         USER=container HOME=/home/container

WORKDIR     /home/container

COPY        ./entrypoint.sh /entrypoint.sh
CMD         [ "/bin/bash", "/entrypoint.sh" ]
