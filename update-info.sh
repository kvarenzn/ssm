#!/usr/bin/bash

wget https://bestdori.com/api/songs/all.5.json -O all.5.json
cat all.5.json | jq | sponge all.5.json
