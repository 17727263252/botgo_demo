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

//func registerForumHandlers(i dto.Intent, handlers ...interface{}) dto.Intent {
//	for _, h := range handlers {
//		switch handle := h.(type) {
//		case ThreadEventHandler:
//			DefaultHandlers.Thread = handle
//			i = i | dto.EventToIntent(
//				dto.EventForumThreadCreate, dto.EventForumThreadUpdate, dto.EventForumThreadDelete,
//			)
//		case PostEventHandler:
//			DefaultHandlers.Post = handle
//			i = i | dto.EventToIntent(dto.EventForumPostCreate, dto.EventForumPostDelete)
//		case ReplyEventHandler:
//			DefaultHandlers.Reply = handle
//			i = i | dto.EventToIntent(dto.EventForumReplyCreate, dto.EventForumReplyDelete)
//		case ForumAuditEventHandler:
//			DefaultHandlers.ForumAudit = handle
//			i = i | dto.EventToIntent(dto.EventForumAuditResult)
//		default:
//		}
//	}
//	return i
//}

// registerRelationHandlers 注册频道关系链相关handlers
//func registerRelationHandlers(i dto.Intent, handlers ...interface{}) dto.Intent {
//	for _, h := range handlers {
//		switch handle := h.(type) {
//		case GuildEventHandler:
//			DefaultHandlers.Guild = handle
//			i = i | dto.EventToIntent(dto.EventGuildCreate, dto.EventGuildDelete, dto.EventGuildUpdate)
//		case GuildMemberEventHandler:
//			DefaultHandlers.GuildMember = handle
//			i = i | dto.EventToIntent(dto.EventGuildMemberAdd, dto.EventGuildMemberRemove, dto.EventGuildMemberUpdate)
//		case ChannelEventHandler:
//			DefaultHandlers.Channel = handle
//			i = i | dto.EventToIntent(dto.EventChannelCreate, dto.EventChannelDelete, dto.EventChannelUpdate)
//		default:
//		}
//	}
//	return i
//}

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
