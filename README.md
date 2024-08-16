# Mess With DNS

The source for [Mess With DNS](https://messwithdns.net).

### Developing

Instructions that are probably missing some important steps:

1. Install powerdns (`apt install pdns-backend-sqlite3 pdns-backend-bind` in Ubuntu, `brew install pdns` in Homebrew)
2. Run `bash run.sh`
3. Open it locally at http://localhost:8080
4. Query the local DNS server with `dig @localhost -p 5354 pear5.messwithdns.com` (replace `pear5` with the domain name that you get when logging in)

### Disclaimers

Two main disclaimers:

1. Probably won't be very actively maintained. I'm very open to bugfix pull
   requests, though I can't guarantee I'll merge them in a timely manner.
2. This code might not be very suitable for running yourself as is, there's a
   bunch of hardcoded stuff ("a wizard zines project") as well as a Honeycomb
   integration for metrics. I'm unlikely to make it easier to run on your own
   but feel free to fork it.
