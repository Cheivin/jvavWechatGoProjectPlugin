package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"wechat-hub-plugin/hub"
)

type Sender struct {
	apiHost  string
	username string
	password string
	sendFn   func(msg hub.SendMsgCommand) error
	client   *http.Client
}

func NewSender(apiHost string, username string, password string, sendFn func(msg hub.SendMsgCommand) error) hub.SenderInterface {
	return &Sender{
		apiHost:  apiHost,
		username: username,
		password: password,
		sendFn:   sendFn,
		client:   &http.Client{},
	}
}
func (s *Sender) SendText(gid string, content string) error {
	return s.sendFn(hub.SendMsgCommand{
		Gid:  gid,
		Type: 1,
		Body: content,
	})
}

func (s *Sender) SendNetworkImg(gid string, src string) error {
	return s.sendFn(hub.SendMsgCommand{
		Gid:  gid,
		Type: 2,
		Body: src,
	})
}

func (s *Sender) SendImg(gid string, filename string, file io.Reader) error {
	src, err := s.upload(filename, file)
	if err != nil {
		slog.Error("Failed to upload image", "error", err)
		return err
	}
	return s.sendFn(hub.SendMsgCommand{
		Gid:      gid,
		Type:     2,
		Body:     src,
		Filename: filename,
	})
}

func (s *Sender) upload(filename string, file io.Reader) (string, error) {
	slog.Info("Uploading image", "filename", filename)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if part, err := writer.CreateFormFile("file", filename); err != nil {
		slog.Error("Failed to create form file", "error", err)
		return "", err
	} else {
		if _, err = io.Copy(part, file); err != nil {
			slog.Error("Failed to copy file", "error", err)
			return "", err
		}
	}

	if err := writer.WriteField("filename", filename); err != nil {
		slog.Error("Failed to write field", "error", err)
		return "", err
	}
	if err := writer.Close(); err != nil {
		slog.Error("Failed to close writer", "error", err)
		return "", err
	}

	req, err := http.NewRequest("POST", apiHost+"/upload", body)
	if err != nil {
		slog.Error("Failed to create request", "error", err)
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(username+":"+password)))

	resp, err := s.client.Do(req)
	if err != nil {
		slog.Error("Failed to send request", "error", err)
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Failed to upload image", "status", resp.Status)
		return "", fmt.Errorf(resp.Status)
	}
	result := httpResult[string]{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		slog.Error("Failed to decode response", "error", err)
		return "", err
	}
	if result.Code != 0 {
		slog.Error("Failed to upload image", "msg", result.Msg)
		return "", fmt.Errorf(result.Msg)
	}
	return result.Data, nil

}
