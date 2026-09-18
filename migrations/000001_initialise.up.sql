CREATE TABLE logs (
  id SERIAL PRIMARY KEY,
  ts TIMESTAMP NOT NULL DEFAULT (now() at time zone 'utc'),
  log_level VARCHAR(20) NOT NULL CHECK(log_level IN ('DEBUG', 'INFO', 'WARNING', 'ERROR')),
  
  -- e.g. login, create_source, etc
  event_type VARCHAR(50) NOT NULL,

  message TEXT NOT NULL,
  extra_data JSONB NOT NULL DEFAULT '{}'::JSONB
);

CREATE TABLE categories (
  id text PRIMARY KEY,
  display_name text,
  cover_image integer DEFAULT -1,
  tags jsonb,
  position integer DEFAULT 900
);

CREATE TABLE sources (
  id serial PRIMARY KEY,
  name text NOT NULL,
  source_type text NOT NULL, -- 'local'
  root_path text NOT NULL UNIQUE,
  enabled boolean NOT NULL DEFAULT true,
  last_scanned_at TIMESTAMPTZ
);

CREATE TABLE path_metadata (
  id serial PRIMARY KEY,
  source_id integer REFERENCES sources(id) ON DELETE CASCADE,
  relative_path text NOT NULL, -- path inside root_path. there's always a relative_path of '' for the whole folder itself
  category_id text REFERENCES categories(id) ON DELETE SET NULL,
  author text,
  author_url text,
  tags jsonb NOT NULL DEFAULT '{}',
  tag_mode text NOT NULL DEFAULT 'merge', -- merge or replace, how to handle parent tags
  UNIQUE (source_id, relative_path)
);

CREATE TABLE images (
  id serial PRIMARY KEY,
  source_id integer REFERENCES sources(id) ON DELETE CASCADE,
  relative_path text NOT NULL, -- e.g. grafitstudio/faces/001.jpg
  file_hash text DEFAULT '',
  local_path text,
  external_url text,

  -- optional metadata overrides, NULL means inherit from folder
  category_override text REFERENCES categories(id) ON DELETE SET NULL,
  author_override text,
  author_url_override text,
  tags_override jsonb DEFAULT NULL, -- NULL = inherit; set value = manual override

  -- calculated final metadata, updated on source scan or folder metadata changes
  effective_category_id text REFERENCES categories(id) ON DELETE SET NULL,
  effective_author text,
  effective_author_url text,
  effective_tags jsonb NOT NULL DEFAULT '{}',

  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (source_id, relative_path)
);

CREATE INDEX idx_images_effective_category ON images(effective_category_id);
CREATE INDEX idx_images_effective_tags ON images USING gin(effective_tags);
