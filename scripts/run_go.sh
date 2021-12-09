set -eu
source scripts/secrets.sh
cd api || exit
go build
cd .. || exit
export DEV=true
export POSTGRES_CONNECTION_STRING='postgres://postgres:mysecretpassword@localhost:5432/postgres?sslmode=disable'
./api/mess-with-dns 5353
