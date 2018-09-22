-- see count connections to db
SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'test_app';
-- clear info about user and db
DROP DATABASE IF EXISTS test_app;
DROP USER IF EXISTS user1;
-- create user and db
CREATE USER user1 WITH PASSWORD 'user1';
CREATE DATABASE test_app;
\connect test_app;
-- create db table
CREATE TABLE "groups" (
	"id" serial NOT NULL,
	"name" VARCHAR(130) NOT NULL,
	"parent_group_id" integer NOT NULL DEFAULT '0'
) WITH (
  OIDS=FALSE
); 
-- give permissions for user on change table
GRANT ALL PRIVILEGES ON groups TO user1;
GRANT ALL PRIVILEGES ON SEQUENCE groups_id_seq TO user1;