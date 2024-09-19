#!/bin/bash


set -m

export REQUEST_DB_FILENAME=/data/requests.sqlite
export USER_DB_FILENAME=/data/users.sqlite
export GOMEMLIMIT=250MiB

# idea from https://doc.powerdns.com/authoritative/performance.html#caches-memory-allocations-glibc
# for reducing powerdns memory usage
export MALLOC_ARENA_MAX=4

pdns_server --config-dir=/etc/pdns &
/usr/bin/mess-with-dns
