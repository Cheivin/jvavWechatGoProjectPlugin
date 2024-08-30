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
	sender      hub.SenderInterface
	pointManage hub.PointInterface
	plugins     []hub.Plugin
}

func NewService(sender hub.SenderInterface, pointManage hub.PointInterface) *Service {
	return &Service{
		sender:      sender,
		pointManage: pointManage,
		plugins:     []hub.Plugin{},
	}
}

func (s *Service) AddPlugin(plugin hub.Plugin) {
	s.plugins = append(s.plugins, plugin)
}

func (s *Service) Handle(message *hub.Message) error {
	slog.Info("receive message", "type", message.MsgType, "content", message.Content)
	ctx := &hub.Context{
		Message: message,
		Sender:  s.sender,
		Point:   s.pointManage,
	}
	for _, plugin := range s.plugins {
		if err := (plugin).Handle(ctx); err != nil {
			return err
		}
		if ctx.IsAbort() {
			break
		}
	}
	return nil
}
