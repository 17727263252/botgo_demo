package sdks

// Package token 用于调用 openapi，websocket 的 token 对象。

// Type token 类型
type Type string

// TokenType
const (
	TypeBot Type = "Bot"
)

// Token 用于调用接口的 token 结构
type Token struct {
	AppID       uint64
	AccessToken string
	Type        Type
}

// NewBotToken 机器人身份的 token
func NewBotToken(appID uint64, accessToken string) *Token {
	return &Token{
		AppID:       appID,
		AccessToken: accessToken,
		Type:        TypeBot,
	}
}
