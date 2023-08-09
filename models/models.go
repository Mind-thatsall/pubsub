package models

import (
	"time"

	"github.com/gocql/gocql"
)

type Subscriber struct {
	UserId     gocql.UUID   `db:"user_id"`
	ServerId   gocql.UUID   `db:"server_id"`
	ChannelsId []gocql.UUID `db:"channels_id"`
}

type Message struct {
	MessageId gocql.UUID `db:"message_id"`
	ChannelId gocql.UUID `db:"channel_id"`
	ServerId  gocql.UUID `db:"server_id"`
	Content   string     `db:"content"`
	SenderId  gocql.UUID `db:"sender_id"`
	CreatedAt time.Time  `db:"created_at"`
}

type Server struct {
	ServerId gocql.UUID `db:"server_id"`
	Name     string     `db:"name"`
}

type Channel struct {
	ChannelId gocql.UUID `db:"channel_id"`
	Name      string     `db:"name"`
	ServerId  gocql.UUID `db:"server_id"`
}
