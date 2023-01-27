#!/bin/bash
# golang generic package
if [ ! -d /mnt/server/ ]; then
mkdir -p /mnt/server/
fi
cd /mnt/server
curl -o main.go https://raw.githubusercontent.com/JewishLewish/PolygonDB/main/main.go
curl -o go.mod https://raw.githubusercontent.com/JewishLewish/PolygonDB/main/go.mod
curl -o go.sum https://raw.githubusercontent.com/JewishLewish/PolygonDB/main/go.sum
mkdir -p /mnt/server/sysadmin/
cd sysadmin
curl -o sysadmin.c https://raw.githubusercontent.com/JewishLewish/PolygonDB/main/sysadmin/sysadmin.c
cd /mnt/server
go get
go build
rm -f main.go
rm -f go.mod
rm -f go.sum
rm -f sysadmin/sysadmin.c
rm -r sysadmin
#settings.json file
if [ ! -f /mnt/server/settings.json ]; then
curl -o settings.json https://raw.githubusercontent.com/JewishLewish/PolygonDB/main/settings.json
fi
#database folder
if [ ! -d /mnt/server/databases ]; then
mkdir -p databases
mkdir -p databases/ExampleDB
cd databases/ExampleDB
curl -o config.json https://raw.githubusercontent.com/JewishLewish/PolygonDB/main/databases/CatoDB/config.json
curl -o database.json https://raw.githubusercontent.com/JewishLewish/PolygonDB/main/databases/CatoDB/database.json
fi
cd /mnt/server