package domain

import "time"

type User struct {
	ID                 string    `bson:"_id,omitempty" json:"id"`
	TelegramID         int64     `bson:"telegram_id" json:"telegram_id"`
	Username           string    `bson:"username" json:"username"`
	FirstName          string    `bson:"first_name" json:"first_name"`
	LastName           string    `bson:"last_name" json:"last_name"`
	LanguageCode       string    `bson:"language_code" json:"language_code"`
	AutoReplyEnabled   bool      `bson:"auto_reply_enabled" json:"auto_reply_enabled"`
	AutoDumpingEnabled bool      `bson:"auto_dumping_enabled" json:"auto_dumping_enabled"` // Глобальный переключатель автодемпинга
	CreatedAt          time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time `bson:"updated_at" json:"updated_at"`
}

type UserRepository interface {
	Create(user *User) error
	GetByTelegramID(telegramID int64) (*User, error)
	Update(user *User) error
	ToggleAutoReply(userID int64, enabled bool) error
	ToggleAutoDumping(userID int64, enabled bool) error
}
