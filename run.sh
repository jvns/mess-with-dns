#!/bin/bash
source secrets.sh
set -eu

pscale connect messwithdns development &
find . -name '*.go'| entr -r bash -c 'cd api; go build && cd .. && PLANETSCALE_CONNECTION_STRING="root:@tcp(localhost:3306)/messwithdns" ./api/mess-with-dns 5353' &
cd frontend
ls *.js components/* | entr bash esbuild.sh &

trap "kill $(jobs -p)" SIGINT SIGTERM

wait
