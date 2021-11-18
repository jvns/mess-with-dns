CREATE TABLE dns_records 
(
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL, -- duplicated just to make it easier to do lookups
  rrtype INTEGER NOT NULL, -- duplicated just to make it easier to do lookups
  content TEXT NOT NULL, -- serialized version of record
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- unique constraint on name, rrtype, content
CREATE UNIQUE INDEX dns_records_name_rrtype_content_uindex ON dns_records (name, rrtype, content);

-- this table will just have one record
CREATE TABLE dns_serials 
(
  serial integer PRIMARY KEY,
);
