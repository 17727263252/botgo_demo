package sdks

import (
	"time"
)

type Intent int

const IntentGuildAtMessage = 1 << 30 // 只接收@消息事件

// ShardConfig 连接的 shard 配置，ShardID 从 0 开始，ShardCount 最小为 1
type ShardConfig struct {
	ShardID    uint32
	ShardCount uint32
}

// Session 连接的 session 结构，包括链接的所有必要字段
type Session struct {
	ID      string
	URL     string
	Token   Token
	Intent  Intent
	LastSeq uint32
	Shards  ShardConfig
}

// NewSession
func NewSession(apiInfo WssConUrlResp, token *Token) Session {
	return Session{
		ID:      time.Now().String(),
		URL:     apiInfo.URL,
		Token:   *token,
		Intent:  IntentGuildAtMessage,
		LastSeq: 0,
		Shards: ShardConfig{
			ShardID:    0,
			ShardCount: apiInfo.Shards,
		},
	}
}

// WSUser 当前连接的用户信息
type WSUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Bot      bool   `json:"bot"`
}
