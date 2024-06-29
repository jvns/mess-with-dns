set -eu
for table in dns_records dns_serials subdomains
do
    echo $table
    bash download_json.sh $table
done
