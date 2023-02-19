#!/bin/bash
# golang generic package
if [ ! -d /mnt/server/ ]; then
mkdir -p /mnt/server/
fi
cd /mnt/server
TAR=v1.3
curl -o main.go https://raw.githubusercontent.com/JewishLewish/PolygonDB/$TAR/main.go
curl -o go.mod https://raw.githubusercontent.com/JewishLewish/PolygonDB/$TAR/go.mod
curl -o go.sum https://raw.githubusercontent.com/JewishLewish/PolygonDB/$TAR/go.sum
cd /mnt/server
go get
go build
rm -f main.go
rm -f go.mod
rm -f go.sum
#settings.json file
if [ ! -f /mnt/server/settings.json ]; then
curl -o settings.json https://raw.githubusercontent.com/JewishLewish/PolygonDB/$TAR/settings.json
fi
#database folder
if [ ! -d /mnt/server/databases ]; then
mkdir -p databases
mkdir -p databases/ExampleDB
cd databases/ExampleDB
curl -o config.json https://raw.githubusercontent.com/JewishLewish/PolygonDB/$TAR/databases/ExampleDB/config.json
curl -o database.json https://raw.githubusercontent.com/JewishLewish/PolygonDB/$TAR/databases/ExampleDB/database.json
fi
cd /mnt/server
