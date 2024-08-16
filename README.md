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

1. There's no license yet, partly beause I don't think this code is very
   suitable for anyone other than me to run, there's a bunch of hardcoded stuff
   ("a wizard zines project") as well as a Honeycomb integration for metrics.
2. Probably won't be very actively maintained. I'm very open to bugfix pull
   requests, though I can't guarantee I'll merge them in a timely manner. I
   have kept the site up for 3 years so far though and I plan to keep it
   running.
