CREATE TABLE IF NOT EXISTS dns_requests 
(
  id INTEGER PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  subdomain VARCHAR(255) NOT NULL,
  src_ip VARCHAR(20) NOT NULL,
  src_host VARCHAR(255) NOT NULL,
  response TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT (strftime('%s','now'))
);

CREATE INDEX IF NOT EXISTS dns_requests_name_uindex ON dns_requests (name);
CREATE INDEX IF NOT EXISTS dns_requests_subdomain_uindex ON dns_requests (subdomain);
CREATE INDEX IF NOT EXISTS dns_requests_created_at ON dns_requests (created_at);
