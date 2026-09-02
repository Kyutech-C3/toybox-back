CREATE TYPE comment_status AS ENUM ('active', 'deleted');
ALTER TABLE comment ADD COLUMN status comment_status NOT NULL DEFAULT 'active';