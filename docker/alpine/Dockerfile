FROM        --platform=$TARGETOS/$TARGETARCH alpine:latest

LABEL       author="JewishLewish/Lenny" maintainer="lenny@lunes.host"

LABEL       org.opencontainers.image.source="https://github.com/JewishLewish/PolygonDB"
LABEL       org.opencontainers.image.licenses=MIT

RUN apk update && apk add --no-cache bash ca-certificates libc6-compat tzdata git
RUN apk add --no-cache libstdc++
RUN apk del --purge *-dev

# Add user with no password, no login and home directory
RUN adduser -D -h /home/container container 

USER container
ENV  HOME /home/container
WORKDIR /home/container

CMD ["/bin/bash"]