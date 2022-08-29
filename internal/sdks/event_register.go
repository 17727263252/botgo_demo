package sdks

// DefaultHandlers 默认的 handlers 结构，管理所有支持的 handlers 类型
var DefaultHandlers struct {
	Ready       ReadyHandler
	ErrorNotify ErrorNotifyHandler
	Plain       PlainEventHandler
	ATMessage   ATMessageEventHandler
}

// ReadyHandler 可以处理 ws 的 ready 事件
type ReadyHandler func(event *WSPayload, data WSReadyData)

// ErrorNotifyHandler 当 ws 连接发生错误的时候，会回调，方便使用方监控相关错误
// 比如 reconnect invalidSession 等错误，错误可以转换为 bot.Err
type ErrorNotifyHandler func(err error)

// PlainEventHandler 透传handler
type PlainEventHandler func(event *WSPayload, message []byte) error

// ATMessageEventHandler at 机器人消息事件 handlers
type ATMessageEventHandler func(event *WSPayload, data *WSATMessageData) error

// RegisterHandlers 注册事件回调，并返回 intent 用于 websocket 的鉴权
func RegisterHandlers(handlers ...interface{}) Intent {
	var i Intent
	for _, h := range handlers {
		switch handle := h.(type) {
		case ReadyHandler:
			DefaultHandlers.Ready = handle
		case ErrorNotifyHandler:
			DefaultHandlers.ErrorNotify = handle
		case PlainEventHandler:
			DefaultHandlers.Plain = handle
		default:
		}
	}
	i = i | registerMessageHandlers(i, handlers...)

	return i
}

//registerMessageHandlers 注册消息相关的 handlers
func registerMessageHandlers(i Intent, handlers ...interface{}) Intent {
	for _, h := range handlers {
		switch handle := h.(type) {
		case ATMessageEventHandler:
			DefaultHandlers.ATMessage = handle
			i = i | 1<<30
		default:
		}
	}
	return i
}
