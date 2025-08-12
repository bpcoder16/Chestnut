package cronserver

type Config struct {
	IsRunCron bool
	CronList  []ConfigItem
}

type ConfigItem struct {
	Name              string
	MaxConcurrencyCnt int

	JobType                 string
	CronJobParams           CronJobParams
	DurationJobParams       DurationJobParams
	DurationRandomJobParams DurationRandomJobParams
}

type CronJobParams struct {
	Crontab string
}

type DurationJobParams struct {
	EveryMillisecond int64
}

type DurationRandomJobParams struct {
	MinMillisecond int64
	MaxMillisecond int64
}
