package LocalDb

import (
	"fmt"
	"log"
	"strings"
)

func InfluxBatchNumbLinss(batch string) int {
	return strings.Count(batch, "\n")
}

func (d SqliteLocalDb) InfluxBacklogAdd(client, batch string) error {
	if err, compressedBatch := compress(batch); err != nil {
		return fmt.Errorf("cannot compress batch: %s", err)
	} else if _, err := d.db.Exec(
		"INSERT INTO influxBacklog (created, client, numbLines, compressedBatch) VALUES(datetime('now'), ?, ?, ?);",
		client,
		InfluxBatchNumbLinss(batch),
		compressedBatch,
	); err != nil {
		return fmt.Errorf("cannot insert into influxBacklog: %s", err)
	}

	return nil
}

func (d SqliteLocalDb) InfluxAggregateBacklog(client string, batchSize uint) error {
	// fetch newest up to 100 rows that sum up to no more batchSize number of lines
	rows, err := d.db.Query(`
SELECT id, numbLines, compressedBatch
FROM influxBacklog
WHERE client = ? AND id >= (
  SELECT MIN(f.id)
    FROM (
     SELECT id, numbLines, SUM(numbLines) OVER (ORDER BY id DESC) AS cum
     FROM influxBacklog
     WHERE client = ?
     ORDER BY id DESC
     LIMIT 16
    ) f
    WHERE f.cum < ?
    GROUP BY NULL
    HAVING COUNT() > 1
);`,
		client, client, batchSize,
	)

	if err != nil {
		return fmt.Errorf(" error during query: %s", err)
	}
	defer rows.Close()

	// aggregate all rows by decompressing,
	var ids []int
	var batches []string
	for rows.Next() {
		var id, numbLines int
		var compressedBatch []byte
		if err := rows.Scan(&id, &numbLines, &compressedBatch); err != nil {
			return fmt.Errorf("error during scan: %s", err)
		}

		ids = append(ids, id)
		if err, batch := uncompress(compressedBatch); err != nil {
			return fmt.Errorf("error during uncompress: %s", err)
		} else {
			batches = append(batches, batch)
		}
	}

	if len(ids) < 1 {
		// nothing to aggregte, free up unused disk space instead
		if d.vacuumNeeded {
			if _, err := d.db.Exec("VACUUM"); err != nil {
				log.Printf("localDb: error during VACUUM: %s", err)
			}
			d.vacuumNeeded = false
		}

		return nil
	}

	// compute new batch
	aggregatedBatch := strings.Join(batches, "")

	// insert aggregated batch
	if err := d.InfluxBacklogAdd(client, aggregatedBatch); err != nil {
		return fmt.Errorf("error during add: %s", err)
	} else {
		// delete old ids that have been aggregated into new batch
		for _, id := range ids {
			if err := d.InfluxBacklogDelete(id); err != nil {
				log.Printf("localDb[%s]: aggregateBacklog: error during delete: %s", client, err)
			}
		}
	}

	log.Printf("localDb[%s]: aggregateBacklog: aggragted %d entries into one", client, len(ids))
	return nil
}

func (d SqliteLocalDb) InfluxBacklogSize(client string) (err error, numbBatches, numbLines uint) {
	row := d.db.QueryRow(
		"SELECT COUNT(*), IFNULL(SUM(numbLines), 0) FROM influxBacklog WHERE client = ?",
		client,
	)
	if e := row.Scan(&numbBatches, &numbLines); e != nil {
		err = fmt.Errorf("cannot select from influxBacklog: %s", e)
	}

	return
}

func (d SqliteLocalDb) InfluxBacklogGet(client string) (err error, id int, batch string) {
	row := d.db.QueryRow(
		"SELECT id, numbLines, compressedBatch FROM influxBacklog WHERE client = ? ORDER BY id ASC LIMIT 1",
		client,
	)
	var numbLines int
	var compressedBatch []byte
	if e := row.Scan(&id, &numbLines, &compressedBatch); e != nil {
		err = fmt.Errorf("cannot select from influxBacklog: %s", e)
	} else {
		err, batch = uncompress(compressedBatch)
		if err != nil {
			err = fmt.Errorf("cannot uncompress: %s", err)
		}
	}

	if count := InfluxBatchNumbLinss(batch); count != numbLines {
		log.Fatalf("numbLines does not match for id=%d, %d != %d", id, count, numbLines)
	}

	return
}

func (d SqliteLocalDb) InfluxBacklogDelete(id int) error {
	if _, err := d.db.Exec("DELETE FROM influxBacklog  WHERE id = ?", id); err != nil {
		return fmt.Errorf("cannot delete from influxBacklog: %s", err)
	}

	d.vacuumNeeded = true

	return nil
}

func (d DisabledLocalDb) InfluxBacklogAdd(client, batch string) error {
	return fmt.Errorf("disabled")
}
func (d DisabledLocalDb) InfluxBacklogSize(client string) (err error, numbBatches, numbLines uint) {
	return fmt.Errorf("disabled"), 0, 0
}
func (d DisabledLocalDb) InfluxBacklogGet(client string) (err error, id int, batch string) {
	return fmt.Errorf("disabled"), 0, ""
}
func (d DisabledLocalDb) InfluxBacklogDelete(id int) error {
	return fmt.Errorf("disabled")
}
func (d DisabledLocalDb) InfluxAggregateBacklog(client string, batchSize uint) error {
	return nil
}
