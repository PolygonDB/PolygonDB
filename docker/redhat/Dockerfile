FROM        --platform=$TARGETOS/$TARGETARCH redhat/ubi8-minimal:8.7

LABEL       author="JewishLewish" \
            maintainer="lenny@lunes.host" \
            org.opencontainers.image.source="https://github.com/JewishLewish/PolygonDB" \
            org.opencontainers.image.licenses=MIT

RUN         rpm --import https://www.redhat.com/security/fd431d51.txt && \
            yum install -y yum

RUN         yum -y update
RUN         yum -y install bash

RUN         yum clean all

USER        container
ENV         USER=container HOME=/home/container

WORKDIR     /home/container

COPY        ./entrypoint.sh /entrypoint.sh
CMD         [ "/bin/bash", "/entrypoint.sh" ]