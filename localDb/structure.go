package LocalDb

const structure string = `
CREATE TABLE dbSchema (
  version INTEGER NOT NULL PRIMARY KEY,
  created DATETIME NOT NULL
);
INSERT INTO dbSchema VALUES (0, datetime('now'));

CREATE TABLE influxBacklog (
  id INTEGER NOT NULL PRIMARY KEY,
  created DATETIME NOT NULL,
  client VARCHAR NOT NULL,
  batch TEXT NOT NULL
);
CREATE UNIQUE INDEX clientIdx ON influxBacklog (client, id);
`
