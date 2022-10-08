package LocalDb

const structure string = `
CREATE TABLE dbSchema (
  version INTEGER PRIMARY KEY NOT NULL,
  created DATETIME NOT NULL
);
INSERT INTO dbSchema VALUES (0, datetime('now'));

CREATE TABLE influxBacklog (
  id INTEGER PRIMARY KEY NOT NULL,
  created DATETIME NOT NULL,
  client VARCHAR NOT NULL,
  numbLines INT NOT NULL,
  compressedBatch BLOB NOT NULL
);
CREATE UNIQUE INDEX clientIdx ON influxBacklog (client, id);
`
