FROM        --platform=$TARGETOS/$TARGETARCH redhat:7

LABEL       author="JewishLewish" \
            maintainer="lenny@lunes.host" \
            org.opencontainers.image.source="https://github.com/JewishLewish/PolygonDB" \
            org.opencontainers.image.licenses=MIT

RUN yum update -y \
    && yum install -y \
        gcc \
        g++ \
        libgcc \
        libstdc++-devel \
        gdb \
        glibc \
        git \
        wget \
        curl \
        tar \
        zip \
        unzip \
        binutils \
        xz \
        lzo \
        cabextract \
        iproute \
        iproute-tc \
        net-tools \
        nc \
        telnet \
        libatomic \
        libSDL1_2 \
        libSDL2_0 \
        fontconfig \
        libicu \
        libicu-de

USER        container
ENV         USER=container HOME=/home/container

WORKDIR     /home/container

COPY        ./entrypoint.sh /entrypoint.sh
CMD         [ "/bin/bash", "/entrypoint.sh" ]
