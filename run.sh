#!/bin/bash
set -eu

ls api/*.go api/go* scripts/* run.sh | entr -r bash scripts/run_go.sh &
cd frontend
ls ./*.js ./*.ts components/* | grep -v bundle | entr bash esbuild.sh &

trap 'kill $(jobs -p)' SIGINT SIGTERM

wait
