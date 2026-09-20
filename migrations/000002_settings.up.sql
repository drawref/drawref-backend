CREATE TABLE IF NOT EXISTS settings (
  key text PRIMARY KEY,
  value text NOT NULL
);

INSERT INTO settings (key, value)
VALUES ('thumbnail_min_filesize_kb', '400')
ON CONFLICT (key) DO NOTHING;
