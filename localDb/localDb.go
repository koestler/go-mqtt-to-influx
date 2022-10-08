package LocalDb

type LocalDb interface {
}

type SqliteLocalDb struct {
	config Config
}

type DisabledLocalDb struct{}

type Config interface {
	Enabled() bool
	Path() string
}

func Run(config Config) LocalDb {
	if config.Enabled() {
		return SqliteLocalDb{
			config: config,
		}
	} else {
		return DisabledLocalDb{}
	}

}
