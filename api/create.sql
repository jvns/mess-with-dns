CREATE TABLE dns_records 
(
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  rrtype INTEGER NOT NULL,
  ttl INTEGER NOT NULL,
  content TEXT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
);
-- unique constraint on name, rrtype, content
CREATE UNIQUE INDEX dns_records_name_rrtype_content_uindex ON dns_records (name, rrtype, content);