set -e

TABLE=$1
printf "\\pset pager off\n;select row_to_json($TABLE) from $TABLE;\nexit\n" | fly pg connect -a mess-with-dns-pg | grep '^ {' | sed 's/^ {/{/' > $TABLE.json

