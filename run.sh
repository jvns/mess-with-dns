#!/bin/bash
set -eu

ls api/*.go api/go* scripts/* run.sh | entr -r bash scripts/run_go.sh &
cd frontend
ls ./*.js ./*.html ./*.ts components/* | grep -v bundle | entr bash esbuild.sh &
npx tailwindcss -o css/tailwind-small.min.css --watch &

trap 'kill $(jobs -p)' SIGINT SIGTERM

wait
