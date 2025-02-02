#!/bin/bash
set -m

export REQUEST_DB_FILENAME=/data/requests.sqlite
export USER_DB_FILENAME=/data/users.sqlite
export GOMEMLIMIT=160MiB
export MALLOC_ARENA_MAX=4

cleanup() {
   pkill -P $$
   exit
}

trap cleanup EXIT

backup() {
    while true; do
        sleep 3h
        echo "Running hourly backup"
        bash /app/backup.sh
    done
}

backup &
pdns_server --config-dir=/etc/pdns &
pdns_pid=$!

/usr/bin/mess-with-dns &
mess_pid=$!

# Wait for either process to exit
wait -n $pdns_pid $mess_pid
