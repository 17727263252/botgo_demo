package handlers

import (
	"fmt"
	"wcxp-project/botgo_demo/internal/sdks"
)

// 自定义业务逻辑，仅需修改相对应的func即可

// Key: channlID+UserID  Value：Timestamp（签到时间戳）
var LocalClockMap = make(map[string]string)

// 打卡
func realDoClockIn(dateStr string, data *sdks.WSATMessageData) (reqMsg sdks.SendMessageReq) {
	mapKey := data.Author.ID + "-" + dateStr + "-" + data.GuildID
	if _, ok := LocalClockMap[mapKey]; ok {
		reqMsg.Content = fmt.Sprintf("<@%s>", data.Author.ID) + fmt.Sprintf("今天已经打过卡啦！")
	} else {
		LocalClockMap[mapKey] = FirstClock
		reqMsg.Content = fmt.Sprintf("<@%s>", data.Author.ID) + fmt.Sprintf(dateStr+"成功打卡！")
	}
	return reqMsg
}

func realDoClockInstruction(dateStr string, data *sdks.WSATMessageData) (reqMsg sdks.SendMessageReq) {
	reqMsg.Content = fmt.Sprintf("1.输入'%s'即可签到,一天仅能签到一次喔！\n2.输入'%s'即可查看本频道下连续签到Top10排行榜,今天你上榜了吗？\n3.输入'%s'即可查看个人签到详情信息喔！", ClockIN, ClockINRanking, ClockINDataSelf)
	return
}

func realDoClockINRanking(dateStr string, data *sdks.WSATMessageData) (reqMsg sdks.SendMessageReq) {
	reqMsg.Content = fmt.Sprintf("功能还未开放！尽请期待！")
	return

}

func realDoClockINDataSelf(dateStr string, data *sdks.WSATMessageData) (reqMsg sdks.SendMessageReq) {
	reqMsg.Content = fmt.Sprintf("功能还未开放！尽请期待！")
	return
}

func realDoInvalidContent(dateStr string, data *sdks.WSATMessageData) (reqMsg sdks.SendMessageReq) {
	reqMsg.Content = fmt.Sprintf("<@%s> 很难过,小胖机器人没有懂这是什么意思呢！", data.Author.ID)
	return
}
