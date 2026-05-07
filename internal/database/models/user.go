package models

type Platform string

const (
	PlatformTelegram Platform = "telegram"
)

type User struct {
	TimeStampedModel

	Platform   Platform `bun:"platform,notnull,type:varchar(10)"`
	PlatformID string   `bun:"platform_id,notnull,index"`
}
