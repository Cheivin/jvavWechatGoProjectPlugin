package main

import (
	"log/slog"
	"wechat-hub-plugin/hub"
)

type httpResult[T any] struct {
	Code int    `json:"code"` // 0表示成功
	Msg  string `json:"msg"`  //
	Data T      `json:"data"`
}

type Service struct {
	sender     *Sender
	plugin_arr []*hub.Plugin
}

func NewService(sender *Sender) *Service {
	return &Service{
		sender:     sender,
		plugin_arr: []*hub.Plugin{},
	}
}

func (s *Service) AddPlugin(plugin *hub.Plugin) {
	s.plugin_arr = append(s.plugin_arr, plugin)
}

func (s *Service) Handle(message *hub.Message) error {
	slog.Info("receive message", "type", message.MsgType, "content", message.Content)
	for _, plugin := range s.plugin_arr {
		if err := (*plugin).Handle(message, s.sender); err != nil {
			return err
		}
	}
	return nil
}
