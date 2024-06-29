mkdir -p sqlite

rm -f sqlite/powerdns.sqlite
sqlite3 sqlite/powerdns.sqlite < ../pdns/pdns.sql
python3 parse_records.py

rm -f  sqlite/users.sqlite

sqlite3 sqlite/users.sqlite < ../api/users/create.sql

sqlite3 sqlite/users.sqlite < ../api/streamer/create.sql
