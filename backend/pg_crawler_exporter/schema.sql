CREATE SCHEMA crawler;

CREATE TABLE crawler.crawl(
    id                  SERIAL PRIMARY KEY,
    timestamp_begin     TIMESTAMP NOT NULL,
    timestamp_end       TIMESTAMP NOT NULL
);

CREATE TABLE crawler.peer(
    id          SERIAL PRIMARY KEY,
    crawl       INTEGER REFERENCES crawler.crawl(id),
    timestamp   TIMESTAMP NOT NULL,
    peer_id     VARCHAR(512) NOT NULL, 
    agent       VARCHAR(512) NOT NULL,
    addresses   VARCHAR(512) ARRAY NOT NULL,
    protocols   VARCHAR(512) ARRAY NOT NULL,
    dht_entries INTEGER NOT NULL,

    ip          VARCHAR(255),
    asn         INTEGER,
    asn_org     VARCHAR(255),
    country     VARCHAR(255),
    city        VARCHAR(255),
    latitude    REAL,
    longitude   REAL,

    CONSTRAINT craw_peer_id_uniq UNIQUE (crawl, peer_id)
);

CREATE INDEX crawler_peer_crawl_index ON crawler.peer USING HASH (crawl);
