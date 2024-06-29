The application has 4 subpackages:

* `db` -- a little wrapper to lock access to the sqlite database
* `users` --  for managing login
* `streamer` -- for streaming DNS requests to the user as they come in, through a websocket
* `records` -- for creating/updating/deleting DNS records (through PowerDNS)
  * `parsing` -- for parsing to/from record

db/db.go

main.go
main_test.go

requestlog/dnstap.go
requestlog/stream.go
requestlog/stream_format.go
requestlog/ip2asn/ip2asn.go

login/oauth.go
login/oauth_test.go
login/users.go
login/users_test.go

records/pdns.go
records/parsing/
