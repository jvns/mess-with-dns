#!/bin/bash
set -eu

cd api || exit
go run schemas/generate_schemas.go > ../frontend/schemas.json
cd .. || exit
ls api/*.go api/go* scripts/* run.sh | entr -r bash scripts/run_go.sh &
sleep 0.5
cd frontend
ls ./*.js ./*.html ./*.ts components/* | grep -v bundle | entr bash esbuild.sh &
npx tailwindcss -o css/tailwind-small.min.css --watch &

trap 'kill $(jobs -p)' SIGINT SIGTERM

wait
