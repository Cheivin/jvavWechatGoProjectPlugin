package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"time"
	"wechat-hub-plugin/hub"
	"wechat-hub-plugin/plugins"
	"wechat-hub-plugin/redirect"
)

var (
	server   string
	apiHost  string
	username string
	password string
)

func init() {
	server = os.Getenv("WS_SERVER")
	username = os.Getenv("WS_USERNAME")
	password = os.Getenv("WS_PASSWORD")
	apiHost = os.Getenv("API_HOST")
	if server == "" {
		panic("WS_SERVER is empty")
	}
	if u, err := url.Parse(server); err != nil {
		panic(err)
	} else {
		query := u.Query()
		query.Set("username", username)
		query.Set("password", password)
		u.RawQuery = query.Encode()
		server = u.String()
	}
}

func initPlugins(service *Service) {
	service.AddPlugin(&plugins.SamePlugin{Model: "realisticVisionV13_v13"})
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	client := redirect.NewWebsocketClientMessageHandler(ctx, server, redirect.WSClientHeartbeat(30*time.Second))

	sender := NewSender(apiHost, username, password, func(msg hub.SendMsgCommand) error {
		command := hub.Command{
			Command: "sendMessage",
			Param:   msg,
		}
		data, err := json.Marshal(command)
		if err != nil {
			slog.Error("命令消息序列化失败", "err", err)
			return err
		}
		return client.SendMessage(data)
	})

	service := NewService(sender)
	initPlugins(service)
	client.OnMessage(func(bs []byte) error {
		message := &hub.Message{}
		if err := json.Unmarshal(bs, message); err != nil {
			slog.Error("消息反序列化失败", "err", err)
			return err
		}
		return service.Handle(message)
	})

	<-ctx.Done()
	defer cancel()
}
