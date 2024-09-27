set -eu
cd api || exit
go build
cd .. || exit
export DEV=true
exec ./api/mess-with-dns 5354
