#!/bin/bash
set -eu

find . -name '*.go'| entr -r bash -c 'cd api; go build; PLANETSCALE_CONNECTION_STRING="root:@tcp(localhost:3306)/messwithdns" ./mess-with-dns 5353' &
cd frontend
echo script.js | entr esbuild script.js  --bundle --minify --outfile=bundle.js &

trap "kill $(jobs -p)" SIGINT SIGTERM

wait
