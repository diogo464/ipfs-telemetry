DROP TABLE IF EXISTS sessions;
CREATE TABLE sessions (
    session         UUID NOT NULL,
    peer            VARCHAR(255) NOT NULL,
    first_seen      TIMESTAMP NOT NULL,
    last_seen       TIMESTAMP NOT NULL,

    PRIMARY KEY (session, peer),
    CHECK (first_seen <= last_seen)
);

DROP TABLE IF EXISTS metrics;
CREATE TABLE metrics (
    session     UUID NOT NULL,
    peer        VARCHAR(255) NOT NULL,
    scope       VARCHAR(255) NOT NULL,
    version     VARCHAR(32) NOT NULL,
    name        VARCHAR(255) NOT NULL,
    attributes  JSONB,
    timestamp   TIMESTAMP NOT NULL,
    value       REAL NOT NULL,

    UNIQUE (session, peer, scope, name, attributes, timestamp)
);

SELECT create_hypertable('metrics', 'timestamp');
CREATE INDEX metrics_session ON metrics (session);
CREATE INDEX metrics_peer ON metrics (peer);
CREATE INDEX metrics_session_peer ON metrics (session,peer);
CREATE INDEX metrics_scope_name ON metrics (scope,name);

DROP TABLE IF EXISTS histograms;
CREATE TABLE histograms (
    session     UUID NOT NULL,
    peer        VARCHAR(255) NOT NULL,
    scope       VARCHAR(255) NOT NULL,
    version     VARCHAR(32) NOT NULL,
    name        VARCHAR(255) NOT NULL,
    attributes  JSONB,
    timestamp   TIMESTAMP NOT NULL,
    count       INTEGER NOT NULL,
    sum         REAL NOT NULL,
    min         REAL NOT NULL,
    max         REAL NOT NULL,
    bounds      REAL[] NOT NULL,
    counts      INTEGER[] NOT NULL,

    UNIQUE(session, peer, scope, name, attributes, timestamp)
);

SELECT create_hypertable('histograms', 'timestamp');
CREATE INDEX histograms_session ON histograms (session);
CREATE INDEX histograms_peer ON histograms (peer);
CREATE INDEX histograms_session_peer ON histograms (session,peer);
CREATE INDEX histograms_scope_name ON histograms (scope,name);

DROP TABLE IF EXISTS properties;
CREATE TABLE properties (
    session     UUID NOT NULL,
    peer        VARCHAR(255) NOT NULL,
    scope       VARCHAR(255) NOT NULL,
    version     VARCHAR(32) NOT NULL,
    name        VARCHAR(255) NOT NULL,
    value       VARCHAR(255) NOT NULL
);

DROP TABLE IF EXISTS events;
CREATE TABLE events (
    session     UUID NOT NULL,
    peer        VARCHAR(255) NOT NULL,
    scope       VARCHAR(255) NOT NULL,
    version     VARCHAR(32) NOT NULL,
    name        VARCHAR(255) NOT NULL,
    timestamp   TIMESTAMP NOT NULL,
    value       JSONB,

    UNIQUE (session, peer, scope, name, timestamp)
);

SELECT create_hypertable('events', 'timestamp');
CREATE INDEX events_session ON events (session);
CREATE INDEX events_peer ON events (peer);
CREATE INDEX events_session_peer ON events (session,peer);
CREATE INDEX events_scope_name ON events (scope,name);

DROP TABLE IF EXISTS discovery;
CREATE TABLE discovery (
    peer        VARCHAR(255),
    timestamp   TIMESTAMP,
    latitude    REAL,
    longitude   REAL,
    location    VARCHAR(255)
);

CREATE INDEX discovery_timestamp ON discovery (peer, timestamp DESC);

CREATE TABLE descriptions (
    scope       VARCHAR(255) NOT NULL,
    version     VARCHAR(32) NOT NULL,
    name        VARCHAR(255) NOT NULL,
    description VARCHAR(255) NOT NULL,
    type        VARCHAR(255) NOT NULL,
    PRIMARY KEY (scope, version, name)
);