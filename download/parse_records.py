import json
import subprocess
from collections import defaultdict
import sqlite3


def get_subdomain(name):
    # 'google.pl.oyster267.messwithdns.com.' -> oyster267.messwithdns.com.
    return ".".join(name.split(".")[-4:])


def group_by_zone(data):
    zones = defaultdict(list)
    for d in data:
        name = d["Hdr"]["Name"]
        zone = get_subdomain(name)
        zones[zone].append(d)
    return zones


def add(d):
    hdr = d["Hdr"]
    name = hdr["Name"]
    zone = get_subdomain(name)
    ttl = str(hdr["Ttl"])
    content = ""
    if hdr["Rrtype"] == 1:
        typ = "A"
        content = d["A"]
    elif hdr["Rrtype"] == 2:
        typ = "NS"
        content = d["Ns"]
    elif hdr["Rrtype"] == 5:
        typ = "CNAME"
        content = d["Target"]
    elif hdr["Rrtype"] == 16:
        typ = "TXT"
        content = '"' + "".join(d["Txt"]) + '"'
    elif hdr["Rrtype"] == 28:
        typ = "AAAA"
        content = d["AAAA"]
    elif hdr["Rrtype"] == 15:
        typ = "MX"
        content = f"{d['Preference']} {d['Mx']}"
    else:
        raise ValueError(f"Unknown record type: {hdr['Rrtype']}")

    subprocess.check_output(
        ["pdnsutil", "--config-dir=.", "add-record", zone, name, typ, ttl, content]
    )


def import_records():
    with open("dns_records.json") as f:
        lines = f.readlines()
        data = [json.loads(line) for line in lines]
        data = [json.loads(r["content"]) for r in data]
    by_zone = group_by_zone(data)
    for zone, records in by_zone.items():
        subprocess.check_output(["pdnsutil", "--config-dir=.", "create-zone", zone])

        for record in records:
            try:
                add(record)
            except Exception as e:
                print(f"Failed to add record: {record}")
                print(e)


def import_users():
    with open("subdomains.json") as f:
        lines = f.readlines()
        data = [json.loads(line) for line in lines]
    conn = sqlite3.connect("sqlite/users.sqlite")
    c = conn.cursor()
    # insert name for each user
    for user in data:
        c.execute("INSERT INTO subdomains (name) VALUES (?)", (user["name"],))
    conn.commit()


def main():
    import_users()
    import_records()


if __name__ == "__main__":
    main()
