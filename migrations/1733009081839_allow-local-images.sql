-- Up Migration
ALTER TABLE images ADD COLUMN is_local boolean default false;

CREATE TABLE local_image_imports (
  id serial primary key,
  path text,
  category_id text,
  author text,
  author_url text,
  tags text[] default '{}'
);

-- Down Migration
ALTER TABLE images DROP COLUMN is_local;

DROP TABLE local_image_imports;
