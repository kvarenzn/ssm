#!/usr/bin/bash

curl https://bestdori.com/api/songs/all.5.json  | jq | sponge all.5.json
