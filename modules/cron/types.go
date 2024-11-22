package cron

type Config struct {
	LockPreName string       `json:"lockPreName"`
	IsRunCron   bool         `json:"isRunCron"`
	CronList    []ConfigItem `json:"cronList"`
}

type ConfigItem struct {
	Name                      string `json:"name"`
	DeadLockExpireMillisecond int64  `json:"deadLockExpireMillisecond"`
	MaxConcurrencyCnt         int    `json:"maxConcurrencyCnt"`

	JobType                 string                  `json:"jobType"`
	CronJobParams           CronJobParams           `json:"cronJobParams"`
	DurationJobParams       DurationJobParams       `json:"durationJobParams"`
	DurationRandomJobParams DurationRandomJobParams `json:"durationRandomJobParams"`
}

type CronJobParams struct {
	Crontab     string `json:"crontab"`
	WithSeconds bool   `json:"withSeconds"`
}

type DurationJobParams struct {
	EveryMillisecond int64 `json:"everyMillisecond"`
}

type DurationRandomJobParams struct {
	MinMillisecond int64 `json:"minMillisecond"`
	MaxMillisecond int64 `json:"maxMillisecond"`
}
