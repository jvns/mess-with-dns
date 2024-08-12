sqlite3 download/sqlite/requests.sqlite <  api/streamer/create.sql

cat download/sqlite/powerdns.sqlite | fly ssh console -C 'bash -c "cat > /data/powerdns.sqlite"'
cat download/sqlite/users.sqlite | fly ssh console -C 'bash -c "cat > /data/users.sqlite"'
cat download/sqlite/requests.sqlite | fly ssh console -C 'bash -c "cat > /data/requests.sqlite"'
