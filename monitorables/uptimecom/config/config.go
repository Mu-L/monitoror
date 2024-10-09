package config

type (
	Uptimecom struct {
		URL             string `validate:"required,url,http"`
		Token           string `validate:"required"`
		Timeout         int    `validate:"gte=0"` // In Millisecond
		CacheExpiration int    `validate:"gte=0"` // In Millisecond
	}
)

var Default = &Uptimecom{
	URL:             "https://uptime.com/api/v1/",
	Token:           "",
	Timeout:         10000,
	CacheExpiration: 30000,
}
