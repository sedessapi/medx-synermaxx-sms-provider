package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"medx-synermaxx-sms-provider/config"
	"medx-synermaxx-sms-provider/pkg/platformhttp"
	"medx-synermaxx-sms-provider/pkg/provider"
	"medx-synermaxx-sms-provider/pkg/synermaxx"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	gin.SetMode(cfg.GinMode)

	client := synermaxx.NewClient(cfg.SynermaxxBaseURL, cfg.SynermaxxAppKey, cfg.SynermaxxOriginator)
	h := provider.NewHandler(client)

	r := gin.Default()
	r.Use(platformhttp.CorrelationMiddleware())
	platformhttp.MountOps(r, "medx-synermaxx-sms-provider", openAPISpec())
	h.RegisterRoutes(r.Group("/provider/v1"))

	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("medx-synermaxx-sms-provider listening on %s (upstream %s)", addr, cfg.SynermaxxBaseURL)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func openAPISpec() map[string]any {
	return map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":       "medx-synermaxx-sms-provider",
			"version":     "v1",
			"description": "Live Synermaxx SMS — Provider Protocol v1. Upstream client from openapi/synermaxx-sendmsg.openapi.yaml.",
		},
		"paths": map[string]any{
			"/healthz":                 map[string]any{"get": map[string]any{"summary": "Liveness"}},
			"/readyz":                  map[string]any{"get": map[string]any{"summary": "Readiness"}},
			"/provider/v1/manifest":    map[string]any{"get": map[string]any{"summary": "Provider manifest"}},
			"/provider/v1/invocations": map[string]any{"post": map[string]any{"summary": "Invoke send-sms"}},
			"/provider/v1/test":        map[string]any{"post": map[string]any{"summary": "Design-time test"}},
		},
	}
}
