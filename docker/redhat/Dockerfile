FROM        --platform=$TARGETOS/$TARGETARCH redhat/ubi8-micro:latest

LABEL       author="JewishLewish" \
            maintainer="lenny@lunes.host" \
            org.opencontainers.image.source="https://github.com/JewishLewish/PolygonDB" \
            org.opencontainers.image.licenses=MIT

RUN curl    -LO https://mirror.rackdc.com/alef/alef/stable/iproute2-5.9.0.tar.xz \
            && tar xf iproute2-5.9.0.tar.xz \
            && cd iproute2-5.9.0 \
            && ./configure \
            && make \
            && make install \
            && cd .. \
            && curl -LO https://ftp.gnu.org/gnu/gawk/gawk-5.1.0.tar.xz \
            && tar xf gawk-5.1.0.tar.xz \
            && cd gawk-5.1.0 \
            && ./configure \
            && make \
            && make install \
            && cd .. \
            && curl -LO https://ftp.gnu.org/gnu/sed/sed-4.8.tar.xz \
            && tar xf sed-4.8.tar.xz \
            && cd sed-4.8 \
            && ./configure \
            && make \
            && make install

USER        container
ENV         USER=container HOME=/home/container

WORKDIR     /home/container

COPY        ./entrypoint.sh /entrypoint.sh
CMD         [ "/bin/bash", "/entrypoint.sh" ]