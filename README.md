# Mess With DNS

The source for [Mess With DNS](https://messwithdns.net).

### Developing

If you want to run it and poke around, here are some instructions that are
probably missing some important steps:

1. Install powerdns (`apt install pdns-backend-sqlite3 pdns-backend-bind` in Ubuntu, `brew install pdns` in Homebrew)
2. Run `bash run.sh`
3. Open it locally at http://localhost:8080
4. Query the local DNS server with `dig @localhost -p 5354 pear5.messwithdns.com` (replace `pear5` with the domain name that you get when logging in)

### Disclaimers

Two main disclaimers:

1. There's no license yet, partly because I don't think this code is very
   suitable for anyone other than me to run, there's a bunch of hardcoded stuff
   like "a wizard zines project" and a metrics integration. Not sure if I'll
   add a license.
2. Probably won't be very actively maintained. I have kept the site up for 3
   years so far though and I plan to keep it running.
