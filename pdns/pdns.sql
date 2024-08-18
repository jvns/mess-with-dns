PRAGMA foreign_keys = 1;
PRAGMA journal_mode=WAL;

CREATE TABLE IF NOT EXISTS domains (
  id                    INTEGER PRIMARY KEY,
  name                  VARCHAR(255) NOT NULL COLLATE NOCASE,
  master                VARCHAR(128) DEFAULT NULL,
  last_check            INTEGER DEFAULT NULL,
  type                  VARCHAR(8) NOT NULL,
  notified_serial       INTEGER DEFAULT NULL,
  account               VARCHAR(40) DEFAULT NULL,
  options               VARCHAR(65535) DEFAULT NULL,
  catalog               VARCHAR(255) DEFAULT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS name_index ON domains(name);
CREATE INDEX IF NOT EXISTS catalog_idx ON domains(catalog);


CREATE TABLE IF NOT EXISTS records (
  id                    INTEGER PRIMARY KEY,
  domain_id             INTEGER DEFAULT NULL,
  name                  VARCHAR(255) DEFAULT NULL,
  type                  VARCHAR(10) DEFAULT NULL,
  content               VARCHAR(65535) DEFAULT NULL,
  ttl                   INTEGER DEFAULT NULL,
  prio                  INTEGER DEFAULT NULL,
  disabled              BOOLEAN DEFAULT 0,
  ordername             VARCHAR(255),
  auth                  BOOL DEFAULT 1,
  FOREIGN KEY(domain_id) REFERENCES domains(id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS records_lookup_idx ON records(name, type);
CREATE INDEX IF NOT EXISTS records_lookup_id_idx ON records(domain_id, name, type);
CREATE INDEX IF NOT EXISTS records_order_idx ON records(domain_id, ordername);


CREATE TABLE IF NOT EXISTS supermasters (
  ip                    VARCHAR(64) NOT NULL,
  nameserver            VARCHAR(255) NOT NULL COLLATE NOCASE,
  account               VARCHAR(40) NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS ip_nameserver_pk ON supermasters(ip, nameserver);


CREATE TABLE IF NOT EXISTS comments (
  id                    INTEGER PRIMARY KEY,
  domain_id             INTEGER NOT NULL,
  name                  VARCHAR(255) NOT NULL,
  type                  VARCHAR(10) NOT NULL,
  modified_at           INT NOT NULL,
  account               VARCHAR(40) DEFAULT NULL,
  comment               VARCHAR(65535) NOT NULL,
  FOREIGN KEY(domain_id) REFERENCES domains(id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS comments_idx ON comments(domain_id, name, type);
CREATE INDEX IF NOT EXISTS comments_order_idx ON comments (domain_id, modified_at);


CREATE TABLE IF NOT EXISTS domainmetadata (
 id                     INTEGER PRIMARY KEY,
 domain_id              INT NOT NULL,
 kind                   VARCHAR(32) COLLATE NOCASE,
 content                TEXT,
 FOREIGN KEY(domain_id) REFERENCES domains(id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS domainmetaidindex ON domainmetadata(domain_id);


CREATE TABLE IF NOT EXISTS cryptokeys (
 id                     INTEGER PRIMARY KEY,
 domain_id              INT NOT NULL,
 flags                  INT NOT NULL,
 active                 BOOL,
 published              BOOL DEFAULT 1,
 content                TEXT,
 FOREIGN KEY(domain_id) REFERENCES domains(id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS domainidindex ON cryptokeys(domain_id);


CREATE TABLE IF NOT EXISTS tsigkeys (
 id                     INTEGER PRIMARY KEY,
 name                   VARCHAR(255) COLLATE NOCASE,
 algorithm              VARCHAR(50) COLLATE NOCASE,
 secret                 VARCHAR(255)
);

CREATE UNIQUE INDEX IF NOT EXISTS namealgoindex ON tsigkeys(name, algorithm);
