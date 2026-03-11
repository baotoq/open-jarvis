// Package main is the entry point for the open-jarvis backend server.
package main

import (
	"flag"
	"net/http"

	"open-jarvis/internal/chat/handler"
	"open-jarvis/internal/config"
	"open-jarvis/internal/config/handler"
	"open-jarvis/internal/conv/handler"
	"open-jarvis/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/config.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	svcCtx := svc.NewServiceContext(c, *configFile)

	server.AddRoute(rest.Route{
		Method:  http.MethodPost,
		Path:    "/api/chat/stream",
		Handler: chat.ChatStreamHandler(svcCtx),
	}, rest.WithSSE())

	server.AddRoutes([]rest.Route{
		{Method: http.MethodPost, Path: "/api/chat/approve", Handler: chat.ApproveHandler(svcCtx)},
		{Method: http.MethodGet, Path: "/api/conversations", Handler: conv.ListConversationsHandler(svcCtx)},
		{Method: http.MethodGet, Path: "/api/conversations/:id", Handler: conv.GetConversationHandler(svcCtx)},
		{Method: http.MethodGet, Path: "/api/conversations/:id/messages", Handler: conv.GetConversationMessagesHandler(svcCtx)},
		{Method: http.MethodDelete, Path: "/api/conversations/:id", Handler: conv.DeleteConversationHandler(svcCtx)},
		{Method: http.MethodGet, Path: "/api/config", Handler: cfg.GetConfigHandler(svcCtx)},
		{Method: http.MethodPut, Path: "/api/config", Handler: cfg.UpdateConfigHandler(svcCtx)},
		{Method: http.MethodGet, Path: "/api/conversations/search", Handler: conv.SearchConversationsHandler(svcCtx)},
	})

	server.Start()
}
