CREATE TABLE fsentry(
  path text NOT NULL UNIQUE 
);

CREATE INDEX path_idx on fsentry (path);
CREATE INDEX path_lower_idx ON fsentry (path COLLATE NOCASE);
