export DEV=true
export WORKDIR=..
export HASH_KEY=CgfCQb/b1yLf251DsG9Zo8CN5h6UKP268QZPxR6ddDw=
export BLOCK_KEY=psYea0IVC59V3kbfMYgWI7AlUmioiNsv9Em1GqksEEE=


cd ../pdns/conf_test || exit 1
rm -f pdns.controlsocket
sqlite3 powerdns.sqlite < ../pdns.sql
pdns_server --config-dir=. &
dnsdist -C dnsdist.conf -l 127.0.0.1:5888 &
cd ../../api || exit 1

# give the servers some time to start
sleep 1

go test ./... "$@"

for job in $(jobs -p); do
    kill -9 $job > /dev/null 2>/dev/null
done

rm -f ../pdns/conf_test/powerdns.sqlite
