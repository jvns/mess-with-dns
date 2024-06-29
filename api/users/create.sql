CREATE TABLE IF NOT EXISTS subdomains (
  name VARCHAR(255) PRIMARY KEY,
  created_at TIMESTAMP NOT NULL DEFAULT (strftime('%s','now'))
);
