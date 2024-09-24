package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"time"
	"wechat-hub-plugin/hub"
	"wechat-hub-plugin/plugins/exit_watch"
	"wechat-hub-plugin/plugins/graph"
	"wechat-hub-plugin/plugins/nga"
	"wechat-hub-plugin/redirect"
)

var (
	server   string
	apiHost  string
	username string
	password string
)

func init() {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")

	viper.SetDefault("PORT", 10000)

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			slog.Error("Config file not found", "err", err)
		} else {
			panic(err)
		}
	}
	viper.AutomaticEnv()

	server = viper.GetString("WS_SERVER")
	username = viper.GetString("WS_USERNAME")
	password = viper.GetString("WS_PASSWORD")
	apiHost = viper.GetString("API_HOST")

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
	// service.AddPlugin(&plugins.SamePlugin{Model: "realisticVisionV13_v13"})
	// service.AddPlugin(write.New())
	service.AddPlugin(exit_watch.Plugin{})
	service.AddPlugin(graph.Plugin{})
	service.AddPlugin(nga.New(os.DirFS(viper.GetString("PLUGIN_NGA_DIR"))))
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

	pointManage := NewPointManage(viper.GetString("API_HOST_POINT"), username, password)

	service := NewService(sender, pointManage)
	service.SetDB(NewDB(connectDB()))

	initPlugins(service)
	client.OnMessage(func(bs []byte) error {
		message := &hub.Message{}
		if err := json.Unmarshal(bs, message); err != nil {
			slog.Error("消息反序列化失败", "err", err)
			return err
		}
		return service.Handle(message)
	})

	go healthEndpoint()
	<-ctx.Done()
	defer cancel()
}

func healthEndpoint() {
	port := viper.GetInt("PORT")
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
		_, _ = w.Write([]byte{})
	})

	slog.Info("HealthEndpoint listening on", "port", port)
	if err := http.ListenAndServe(":"+strconv.Itoa(port), mux); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("HealthEndpoint ListenAndServe", "err", err)
	}
}

func connectDB() *sql.DB {
	host := viper.GetString("DB_HOST")
	port := viper.GetInt("DB_PORT")
	username := viper.GetString("DB_USERNAME")
	password := viper.GetString("DB_PASSWORD")
	database := viper.GetString("DB_DATABASE")
	parameter := viper.GetString("DB_PARAMETER")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", username, password, host, port, database, parameter)
	// 打开数据库连接
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(5)
	slog.Info("Database connected")
	return db
}
