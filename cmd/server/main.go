package main

import (
	"botgo_demo/internal/handlers"
	"botgo_demo/internal/sdks"
	"botgo_demo/internal/sdks/manager"
	"context"
	"log"
	"time"
)

const (
	// todo: 配置文件
	APPID = 102022260
	TOKEN = "QbDX7E2DN06cSra5Oj0QaXHdbP4IS7d5"
)

func main() {
	logger := log.Logger{}

	ctx := context.Background()

	_, cancelFunc := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
	defer func() {
		cancelFunc()
	}()
	// 获取HTTP请求客户端
	httpClient := sdks.NewDefaultHttpClient()
	// 获取自定義封裝的API客户端
	apiClient := sdks.NewApiClient(httpClient)
	// 获取websocket请求的url
	wssConUrlResp, err := apiClient.GetWssConUrl(ctx)
	if err != nil {
		logger.Println("[Get_Wss_Con_Url_Fail],err:", err)
		return
	}
	newToken := sdks.NewBotToken(APPID, TOKEN)
	chanManager := manager.NewChanManager()
	intent := sdks.RegisterHandlers(handlers.ATMessageEventHandler(context.TODO(), apiClient))
	err = chanManager.Start(&wssConUrlResp, newToken, intent)
	if err != nil {
		logger.Println("[Manager_Start_Fail],err:", err)
		return
	}
}
