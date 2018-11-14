DROP TABLE votes CASCADE;
DROP TABLE users CASCADE;
DROP TABLE forums CASCADE;
DROP TABLE posts CASCADE;
DROP TABLE threads CASCADE;

CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS users (
	about CITEXT,
	email CITEXT NOT NULL UNIQUE,
	fullname CITEXT NOT NULL,
	nickname CITEXT COLLATE "ucs_basic" NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS forums (
	posts BIGINT DEFAULT 0,
	slug CITEXT NOT NULL UNIQUE,
	threads INTEGER DEFAULT 0,
	title CITEXT NOT NULL,
	author CITEXT COLLATE "ucs_basic" NOT NULL REFERENCES users (nickname)
);

CREATE TABLE IF NOT EXISTS threads (
	id SERIAL PRIMARY KEY,
	author CITEXT COLLATE "ucs_basic" NOT NULL REFERENCES users (nickname),
	created TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp,
	forum CITEXT NOT NULL REFERENCES forums (slug), 
	message CITEXT NOT NULL,
	slug CITEXT UNIQUE,
	title CITEXT NOT NULL,
	votes INTEGER DEFAULT 0
);

CREATE TABLE IF NOT EXISTS posts (
	author CITEXT COLLATE "ucs_basic" NOT NULL REFERENCES users (nickname),
	created TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp,
	forum CITEXT REFERENCES forums (slug),
	id BIGSERIAL PRIMARY KEY,
	isedited BOOLEAN DEFAULT FALSE,
	message text NOT NULL,
	parent BIGINT DEFAULT 0,
	thread INTEGER NOT NULL REFERENCES threads (id)
);

CREATE TABLE IF NOT EXISTS votes (
	nickname CITEXT COLLATE "ucs_basic" NOT NULL REFERENCES users (nickname),
	voice INTEGER NOT NULL,
	thread INTEGER NOT NULL REFERENCES threads (id),
	UNIQUE (nickname, thread)
);

GRANT ALL PRIVILEGES ON ALL TABLEs IN schema public to tpforumsapi;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO tpforumsapi;

CREATE OR REPLACE FUNCTION check_message() RETURNS TRIGGER AS '
BEGIN
  NEW.isedited:=false;
  RETURN NEW;
END;
'
LANGUAGE plpgsql;

CREATE TRIGGER change_message
BEFORE UPDATE ON posts FOR EACH ROW WHEN (new.message=old.message)
EXECUTE PROCEDURE check_message();


CREATE OR REPLACE FUNCTION post_create() RETURNS TRIGGER AS '
BEGIN
  UPDATE forums SET posts=posts+1 WHERE slug=NEW.forum;
  RETURN NEW;
END;
'
LANGUAGE plpgsql;

CREATE TRIGGER post_create
BEFORE INSERT ON posts FOR EACH ROW
EXECUTE PROCEDURE post_create();


CREATE OR REPLACE FUNCTION thread_create() RETURNS TRIGGER AS '
BEGIN
  UPDATE forums SET threads=threads+1 WHERE slug=NEW.forum;
  RETURN NEW;
END;
'
LANGUAGE plpgsql;

CREATE TRIGGER thread_create
BEFORE INSERT ON threads FOR EACH ROW
EXECUTE PROCEDURE thread_create();

