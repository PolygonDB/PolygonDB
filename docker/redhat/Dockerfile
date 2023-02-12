FROM        --platform=$TARGETOS/$TARGETARCH redhat/ubi8-minimal:latest

LABEL       author="JewishLewish" \
            maintainer="lenny@lunes.host" \
            org.opencontainers.image.source="https://github.com/JewishLewish/PolygonDB" \
            org.opencontainers.image.licenses=MIT

# Get a list of all installed packages
RUN         microdnf list installed | awk '{ if ($1 != "microdnf") print $1 }' > /tmp/packages

# Remove all packages except for microdnf
RUN         while read package; do microdnf remove "$package"; done < /tmp/packages


USER        container
ENV         USER=container HOME=/home/container

WORKDIR     /home/container

COPY        ./entrypoint.sh /entrypoint.sh
CMD         [ "/bin/bash", "/entrypoint.sh" ]