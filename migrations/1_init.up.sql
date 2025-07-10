CREATE TABLE IF NOT EXISTS main
(
    id           SERIAL PRIMARY KEY,
    name         TEXT NOT NULL UNIQUE,
    phone_number TEXT NOT NULL,
    email        TEXT  UNIQUE
);

SET search_path TO grafit;

CREATE TABLE IF NOT EXISTS workers
(
    id           SERIAL PRIMARY KEY,
    surname      TEXT NOT NULL,
    name         TEXT NOT NULL,
    middle_name  TEXT,
    email        TEXT NOT NULL UNIQUE,
    phone_number TEXT NOT NULL,
    cabinet      TEXT,
    position     TEXT,
    department   TEXT NOT NULL,
    birth_date   DATE,
    description  TEXT,
    photo        BYTEA
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
    cabinet      TEXT,
    position     TEXT,
    department   TEXT NOT NULL,
    birth_date   DATE,
    description  TEXT,
    photo        BYTEA
);