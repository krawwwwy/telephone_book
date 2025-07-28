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
    section     TEXT,
    birth_date   DATE,
    description  TEXT,
    photo        BYTEA
);

CREATE TABLE IF NOT EXISTS departments 
(
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
);

create table if not exists sections
(
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    parent_id INT NULL,
    FOREIGN KEY (parent_id) REFERENCES departments(id) ON DELETE CASCADE
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
    section     TEXT,
    birth_date   DATE,
    description  TEXT,
    photo        BYTEA
);

CREATE TABLE IF NOT EXISTS departments 
(
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
);

create table if not exists sections
(
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    parent_id INT NULL,
    FOREIGN KEY (parent_id) REFERENCES departments(id) ON DELETE CASCADE
);