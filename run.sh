#!/bin/bash
set -eu

pscale connect messwithdns development &
find . -name '*.go'| entr -r bash -c 'cd api; go build; cd ..; PLANETSCALE_CONNECTION_STRING="root:@tcp(localhost:3306)/messwithdns" ./api/mess-with-dns 5353' &
cd frontend
echo script.js | entr esbuild script.js  --sourcemap --bundle --minify --outfile=bundle.js &

trap "kill $(jobs -p)" SIGINT SIGTERM

wait
