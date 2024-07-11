export POSTGRES_CONNECTION_STRING='postgres://postgres:mysecretpassword@localhost:5432/postgres?sslmode=disable'
export DEV=true
export WORKDIR=..
export HASH_KEY=CgfCQb/b1yLf251DsG9Zo8CN5h6UKP268QZPxR6ddDw=
export BLOCK_KEY=psYea0IVC59V3kbfMYgWI7AlUmioiNsv9Em1GqksEEE=
go test ./... "$@"
