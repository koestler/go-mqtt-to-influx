package LocalDb

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

type LocalDb interface {
	Enabled() bool
	Shutdown()
	InfluxBacklogAdd(client, batch string) error
	InfluxBacklogSize(client string) (numbBatches, numbLines uint, err error)
	InfluxBacklogGet(client string) (id int, batch string, err error)
	InfluxBacklogDelete(id int) error
	InfluxAggregateBacklog(client string, batchSize uint) error
}

type SqliteLocalDb struct {
	config       Config
	db           *sql.DB
	vacuumNeeded bool
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

			return &SqliteLocalDb{
				config:       config,
				db:           db,
				vacuumNeeded: true,
			}
		}
	}

	return &DisabledLocalDb{}
}

func (d SqliteLocalDb) Enabled() bool {
	return true
}

func (d SqliteLocalDb) Shutdown() {
	if err := d.db.Close(); err != nil {
		log.Printf("localDb: error during close: %s", err)
	} else {
		log.Print("localDb: closed")
	}
}

func (d DisabledLocalDb) Enabled() bool {
	return false
}
func (d DisabledLocalDb) Shutdown() {}
