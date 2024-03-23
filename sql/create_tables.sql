CREATE DATABASE docker;
\c docker;
CREATE ROLE postgres SUPERUSER LOGIN PASSWORD 'postgres';
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE IF NOT EXISTS jarmuvek (
  uuid UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
  rendszam VARCHAR(30) NULL,
  tulajdonos VARCHAR(50) NULL,
  forgalmi_ervenyes DATE NULL,
  adatok TEXT[]
);
CREATE INDEX ON jarmuvek (uuid);
CLUSTER jarmuvek USING jarmuvek_uuid_idx;