#!/usr/bin/bash

wget https://bestdori.com/api/songs/all.5.json -O 5.json
cat 5.json | jq > all.5.json
rm 5.json
