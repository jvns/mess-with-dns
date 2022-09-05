CREATE TABLE IF NOT EXISTS dns_records 
(
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL, -- duplicated just to make it easier to do lookups
  subdomain VARCHAR(255) NOT NULL, -- duplicated just to make it easier to do lookups
  rrtype INTEGER NOT NULL, -- duplicated just to make it easier to do lookups
  content TEXT NOT NULL, -- serialized version of record
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS dns_records_name_rrtype_content_uindex ON dns_records (name, rrtype, content);
CREATE INDEX IF NOT EXISTS dns_records_subdomain ON dns_records (subdomain);

CREATE TABLE IF NOT EXISTS dns_serials 
(
  serial integer PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS dns_requests 
(
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  subdomain VARCHAR(255) NOT NULL,
  src_ip VARCHAR(20) NOT NULL,
  src_host VARCHAR(255) NOT NULL,
  request TEXT NOT NULL,
  response TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS subdomains (
  name VARCHAR(255) PRIMARY KEY,
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS dns_requests_name_uindex ON dns_requests (name);
CREATE INDEX IF NOT EXISTS dns_requests_subdomain_uindex ON dns_requests (subdomain);
