create table snapshots_rt (
    snapshot_rt_id      INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    snapshot_time       TIMESTAMP NOT NULL,
    session_id          INT NOT NULL,
    buckets             INT[] NOT NULL,

    CONSTRAINT fk_sessions FOREIGN KEY (session_id) REFERENCES sessions (session_id)
);
