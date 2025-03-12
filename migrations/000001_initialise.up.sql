CREATE TABLE categories (
  id text primary key,
  display_name text,
  cover_image integer default -1,
  tags jsonb,
  position integer default 900
);

CREATE TABLE images (
  -- identification of the image
  id serial primary key,
  image_hash text default '',

  -- each image has one of these
  s3_path text,
  local_path text,
  external_url text,

  -- metadata
  author text,
  author_url text
);

CREATE TABLE image_tags (
  category_id text references categories(id),
  image_id integer references images(id),
  tags text[] not null default '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  primary key (category_id, image_id)
);

CREATE TABLE local_image_folders (
  local_path text primary key
);
