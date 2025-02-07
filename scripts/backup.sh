#!/bin/bash
set -eux

rm -f /tmp/*.sqlite.gz /tmp/*.sqlite

# Backup & compress our database to /tmp
#
sqlite3 /data/powerdns.sqlite "VACUUM INTO '/tmp/powerdns.sqlite'"
sqlite3 /data/users.sqlite "VACUUM INTO '/tmp/users.sqlite'"

gzip /tmp/powerdns.sqlite
gzip /tmp/users.sqlite

# Upload backup to S3
export RESTIC_PASSWORD='notasecret'
# try to reduce memory usage
export GOGC=20
restic -r s3://s3.amazonaws.com/wizardzines-db-backup/messwithdns/ backup /tmp/powerdns.sqlite.gz /tmp/users.sqlite.gz
restic -r s3://s3.amazonaws.com/wizardzines-db-backup/messwithdns/ snapshots
restic -r s3://s3.amazonaws.com/wizardzines-db-backup/messwithdns/ forget -l 1 -H 6 -d 2 -w 2 -m 2 -y 2
restic -r s3://s3.amazonaws.com/wizardzines-db-backup/messwithdns/ prune

# Notify dead man that back up completed successfully.
curl https://hc-ping.com/fbe4a2e0-245c-4815-aa2e-9c0f97ce308e
