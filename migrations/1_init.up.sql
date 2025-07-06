SET search_path TO grafit;

CREATE TABLE IF NOT EXISTS workers
(
    id           SERIAL PRIMARY KEY,
    surname      TEXT NOT NULL,
    name         TEXT NOT NULL,
    middle_name  TEXT,
    email        TEXT NOT NULL UNIQUE,
    phone_number TEXT NOT NULL,
    cabinet      TEXT NOT NULL,
    position     TEXT NOT NULL,
    department   TEXT NOT NULL
);

SET search_path TO giredmet;

CREATE TABLE IF NOT EXISTS workers
(
    id           SERIAL PRIMARY KEY,
    surname      TEXT NOT NULL,
    name         TEXT NOT NULL,
    middle_name  TEXT,
    email        TEXT NOT NULL UNIQUE,
    phone_number TEXT NOT NULL,
    cabinet      TEXT NOT NULL,
    position     TEXT NOT NULL,
    department   TEXT NOT NULL
);