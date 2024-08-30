package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"wechat-hub-plugin/hub"
)

type (
	PointManage struct {
		apiHost  string
		username string
		password string
		client   *http.Client
	}

	payPoint struct {
		GID     string `json:"gid"`
		UID     string `json:"uid"`
		Point   int    `json:"point"`
		Command string `json:"command"`
	}
	pointResult struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data int    `json:"data"`
	}
)

func NewPointManage(apiHost string, username string, password string) hub.PointInterface {
	return &PointManage{
		apiHost:  apiHost,
		username: username,
		password: password,
		client:   &http.Client{},
	}
}

func (p PointManage) Pay(gid string, uid string, point int, command string) (int, error) {
	data := payPoint{
		GID:     gid,
		UID:     uid,
		Point:   point,
		Command: command,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		slog.Error("Error marshaling JSON", "data", data, "error", err)
		return 0, fmt.Errorf("组装请求失败")
	}
	req, err := http.NewRequestWithContext(context.Background(), "POST", p.apiHost+"/api/point/deduction/command", bytes.NewBuffer(jsonData))
	if err != nil {
		slog.Error("Error creating request", "error", err)
		return 0, fmt.Errorf("创建请求失败")
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		slog.Error("Error sending request", "error", err)
		return 0, fmt.Errorf("发送请求失败")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	result := new(pointResult)
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		slog.Error("Error reading response body", "error", err)
		return 0, fmt.Errorf("请求失败")
	}
	if result.Code != 0 {
		slog.Error("pay point failed", "gid", gid, "uid", uid, "point", strconv.Itoa(point), "command", command, "code", result.Code, "msg", result.Msg)
		return result.Code, fmt.Errorf(result.Msg)
	}
	slog.Info("pay point", "gid", gid, "uid", uid, "point", strconv.Itoa(point), "command", command, "result", result)
	return result.Data, nil
}
