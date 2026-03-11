package main

import (
	"flag"
	"net/http"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"open-jarvis/internal/config"
	"open-jarvis/internal/handler"
	"open-jarvis/internal/svc"
)

var configFile = flag.String("f", "etc/config.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	svcCtx := svc.NewServiceContext(c)

	server.AddRoute(rest.Route{
		Method:  http.MethodPost,
		Path:    "/api/chat/stream",
		Handler: handler.ChatStreamHandler(svcCtx),
	}, rest.WithSSE())

	server.AddRoutes([]rest.Route{
		{Method: http.MethodPost, Path: "/api/chat/approve", Handler: handler.ApproveHandler(svcCtx)},
		{Method: http.MethodGet, Path: "/api/conversations", Handler: handler.ListConversationsHandler(svcCtx)},
		{Method: http.MethodGet, Path: "/api/conversations/:id", Handler: handler.GetConversationHandler(svcCtx)},
		{Method: http.MethodGet, Path: "/api/conversations/:id/messages", Handler: handler.GetConversationMessagesHandler(svcCtx)},
		{Method: http.MethodDelete, Path: "/api/conversations/:id", Handler: handler.DeleteConversationHandler(svcCtx)},
	})

	server.Start()
}
