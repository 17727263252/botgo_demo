package main

import (
	"context"
	"log"
	"time"
	"wcxp-project/botgo_demo/config"
	"wcxp-project/botgo_demo/internal/handlers"
	"wcxp-project/botgo_demo/internal/sdks"
	"wcxp-project/botgo_demo/internal/sdks/manager"
)

func main() {
	logger := log.Logger{}
	ctx := context.Background()
	_, cancelFunc := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
	defer func() {
		cancelFunc()
	}()
	// 1. 初始化配置 appid\token等
	if err := config.Init(); err != nil {
		panic(err)
	}
	// 2. 初始化HTTP请求客户端
	httpClient := sdks.NewDefaultHttpClient()
	// 3. 初始化自定义封裝的API客户端
	apiClient := sdks.NewApiClient(httpClient)
	// 4. 获取websocket建连地址
	wssConUrlResp, err := apiClient.GetWssConUrl(ctx)
	if err != nil {
		logger.Println("[Get_Wss_Con_Url_Fail],err:", err)
		return
	}
	// 5. 初始化会话管理
	newToken := sdks.NewBotToken(config.Conf.APPID, config.Conf.TOKEN)
	chanManager := manager.NewChanManager()
	// 6. 初始化监听的事件, 并注册相应的回调函数
	intent := sdks.RegisterHandlers(handlers.ATMessageEventHandler(context.TODO(), apiClient))
	// 7. 发起实际的websocket请求
	err = chanManager.Start(&wssConUrlResp, newToken, intent)
	if err != nil {
		logger.Println("[Manager_Start_Fail],err:", err)
		return
	}
}
