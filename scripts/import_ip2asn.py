# /// scripq
# requires-python = ">=3.9"
# dependencies = [
#   "sqlite-utils",
# ]
# ///
import sqlite_utils
import csv
import ipaddress


def ip_to_int(ip):
    return int(ipaddress.ip_address(ip))


def parse_row_ipv4(row):
    name = row[3]
    if " - " in name:
        name = name.split(" - ")[0]
    return {
        "start_ip": ip_to_int(row[0]),
        "end_ip": ip_to_int(row[1]),
        "asn": int(row[2]),
        "name": name,
        "country": row[4],
    }


def parse_row_ipv6(row):
    name = row[3]
    if " - " in name:
        name = name.split(" - ")[3]
    return {
        "start_ip": ipaddress.ip_address(row[0]).exploded,
        "end_ip": ipaddress.ip_address(row[1]).exploded,
        "asn": int(row[2]),
        "name": name,
        "country": row[4],
    }


def main():
    db = sqlite_utils.Database("ip-ranges.sqlite")
    db["ipv4_ranges"].create(
        {"start_ip": int, "end_ip": int, "asn": int, "country": str, "name": str},
        if_not_exists=True,
        not_null=["start_ip", "end_ip", "asn", "country", "name"],
    )

    db["ipv6_ranges"].create(
        {"start_ip": str, "end_ip": str, "asn": int, "country": str, "name": str},
        if_not_exists=True,
        not_null=["start_ip", "end_ip"],
    )
    db["ipv4_ranges"].create_index(["start_ip"], if_not_exists=True)
    db["ipv6_ranges"].create_index(["start_ip"], if_not_exists=True)

    with open("ip2asn-v4.tsv") as f:
        reader = csv.reader(f, delimiter="\t")
        data = [parse_row_ipv4(row) for row in reader]
        db["ipv4_ranges"].insert_all(data)

    with open("ip2asn-v6.tsv") as f:
        reader = csv.reader(f, delimiter="\t")
        data = [parse_row_ipv6(row) for row in reader]
        db["ipv6_ranges"].insert_all(data)


if __name__ == "__main__":
    main()
