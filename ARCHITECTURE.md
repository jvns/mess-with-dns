The application has 4 subpackages:

* `db` -- a little wrapper to lock access to the sqlite database
* `users` --  for managing login
* `streamer` -- for streaming DNS requests to the user as they come in, through a websocket
* `records` -- for creating/updating/deleting DNS records (through PowerDNS)
  * `parsing` -- for parsing to/from record
