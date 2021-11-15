#!/bin/bash
set -eu
cd api
go build
cd ..


pscale connect messwithdns development &
# get pid
find . -name '*.go'| entr -r env PLANETSCALE_CONNECTION_STRING="root:@tcp(localhost:3306)/messwithdns" ./api/mess-with-dns 5353 &

cd frontend
echo script.js | entr esbuild script.js  --bundle --minify --outfile=bundle.js &

trap "kill $(jobs -p)" SIGINT SIGTERM

wait
