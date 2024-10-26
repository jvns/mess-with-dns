#!/bin/bash


set -m

export REQUEST_DB_FILENAME=/data/requests.sqlite
export USER_DB_FILENAME=/data/users.sqlite

export GOMEMLIMIT=160MiB

# idea from https://doc.powerdns.com/authoritative/performance.html#caches-memory-allocations-glibc
# for reducing powerdns memory usage
export MALLOC_ARENA_MAX=4

backup() {
    while true; do
        sleep 3h
        echo "Running hourly backup"
        bash /app/backup.sh
    done
}

backup &
pdns_server --config-dir=/etc/pdns &
/usr/bin/mess-with-dns
