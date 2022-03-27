create table sessions (
    session_id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    session_uuid uuid NOT NULL,
    peerid varchar(64) NOT NULL,
    last_seen TIMESTAMP NOT NULL,

    CONSTRAINT uuid_pid_uniq UNIQUE (session_uuid, peerid)
);
