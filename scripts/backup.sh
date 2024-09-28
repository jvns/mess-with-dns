#!/bin/bash
set -eux

# Backup & compress our database to the temp directory.

sqlite3 /data/powerdns.sqlite "VACUUM INTO '/tmp/powerdns.sqlite'"
sqlite3 /data/users.sqlite "VACUUM INTO '/tmp/users.sqlite'"
sqlite3 /data/requests.sqlite "VACUUM INTO '/tmp/requests.sqlite'"

gzip /tmp/powerdns.sqlite
gzip /tmp/users.sqlite
gzip /tmp/requests.sqlite

# Upload backup to S3
export RESTIC_PASSWORD='notasecret'
restic -r s3://s3.amazonaws.com/wizardzines-db-backup/messwithdns/ backup /tmp/powerdns.sqlite.gz /tmp/requests.sqlite.gz /tmp/users.sqlite.gz
restic -r s3://s3.amazonaws.com/wizardzines-db-backup/messwithdns/ snapshots
restic -r s3://s3.amazonaws.com/wizardzines-db-backup/messwithdns/ forget -l 7 -H 12 -d 2 -w 2 -m 2 -y 2
restic -r s3://s3.amazonaws.com/wizardzines-db-backup/messwithdns/ prune

# Notify dead man that back up completed successfully.
curl https://hc-ping.com/fbe4a2e0-245c-4815-aa2e-9c0f97ce308e
