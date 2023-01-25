FROM        --platform=$TARGETOS/$TARGETARCH alpine:latest

LABEL       author="Michael Parker" maintainer="parker@pterodactyl.io"

LABEL       org.opencontainers.image.source="https://github.com/JewishLewish/PolygonDB"
LABEL       org.opencontainers.image.licenses=MIT

RUN         apk update && apk upgrade -U -a \
                && apk add gcc g++ libgcc libstdc++ gdb libc-dev git wget curl tar zip unzip binutils xz-libs lzo-dev cabextract iproute2 net-tools netcat-openbsd telnet libatomic libsdl1.2 libsdl2-dev \
                libfontconfig libicu-dev icu-devtools libunwind libssl1.1-dev sqlite-dev mariadb-dev-libs duktape liblua5.3-dev zlib-dev rapidjson-dev \
                && echo "en_US.UTF-8 UTF-8" >> /etc/locale.gen \
                && locale-gen en_US.UTF-8 \
                && adduser -D -h /home/container -s /bin/ash container
