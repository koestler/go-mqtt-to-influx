package LocalDb

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

type LocalDb interface {
	Shutdown()
	InfluxBacklogAdd(client, batch string) error
	InfluxBacklogGet(client string) (err error, id int, batch string)
	InfluxBacklogDelete(id int) error
}

type SqliteLocalDb struct {
	config Config
	db     *sql.DB
}

type DisabledLocalDb struct{}

type Config interface {
	Enabled() bool
	Path() string
}

func Run(config Config) LocalDb {
	if config.Enabled() {
		db, err := sql.Open("sqlite3", config.Path())
		if err != nil {
			log.Printf("localDb: cannot start sqlite3 db: %s", err)
		} else {
			row := db.QueryRow("SELECT MAX(version) FROM dbSchema")
			var version int
			if err := row.Scan(&version); err != nil {
				// create schema
				if _, err := db.Exec(structure); err != nil {
					log.Printf("localDb: error while creating db structure: %s", err)
				} else {
					log.Printf("localDb: db initialized with schema version 0")
				}
			} else {
				log.Printf("localDb: db schema up-to-date at version: %d", version)
			}

			return SqliteLocalDb{
				config: config,
				db:     db,
			}
		}
	}

	return DisabledLocalDb{}
}

func (d SqliteLocalDb) Shutdown() {
	if err := d.db.Close(); err != nil {
		log.Printf("localDb: error during close: %s", err)
	} else {
		log.Print("localDb: closed")
	}
}

func (d SqliteLocalDb) InfluxBacklogAdd(client, batch string) error {
	if _, err := d.db.Exec(
		"INSERT INTO influxBacklog (created, client, batch) VALUES(datetime('now'), ?, ?);",
		client,
		batch,
	); err != nil {
		return fmt.Errorf("cannot insert into influxBacklogV0: %s", err)
	}

	return nil
}

func (d SqliteLocalDb) InfluxBacklogGet(client string) (err error, id int, batch string) {
	row := d.db.QueryRow(
		"SELECT id, batch FROM influxBacklog WHERE client = ? ORDER BY id ASC LIMIT 1",
		client,
	)
	if e := row.Scan(&id, &batch); e != nil {
		err = fmt.Errorf("cannot select frominfluxBacklogV0: %s", e)
	}

	return
}

func (d SqliteLocalDb) InfluxBacklogDelete(id int) error {
	if _, err := d.db.Exec("DELETE FROM influxBacklog  WHERE id = ?", id); err != nil {
		return fmt.Errorf("cannot delete from influxBacklogV0: %s", err)
	}

	return nil
}

func (d DisabledLocalDb) Shutdown() {}
func (d DisabledLocalDb) InfluxBacklogAdd(client, batch string) error {
	return fmt.Errorf("disabled")
}
func (d DisabledLocalDb) InfluxBacklogGet(client string) (err error, id int, batch string) {
	return fmt.Errorf("disabled"), 0, ""
}
func (d DisabledLocalDb) InfluxBacklogDelete(id int) error {
	return fmt.Errorf("disabled")
}
