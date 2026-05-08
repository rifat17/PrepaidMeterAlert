package models

import "github.com/uptrace/bun"

type Platform string

const (
	PlatformTelegram Platform = "telegram"
)

type User struct {
	bun.BaseModel `bun:"table:users"`

	TimeStampedModel

	Platform   Platform `bun:"platform,notnull,type:varchar(10)"`
	PlatformID string   `bun:"platform_id,notnull,type:varchar(20)"`
}
