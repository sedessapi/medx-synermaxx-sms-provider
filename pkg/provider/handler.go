package provider

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"medx-synermaxx-sms-provider/pkg/synermaxx"
)

type Handler struct {
	Client *synermaxx.Client
}

func NewHandler(client *synermaxx.Client) *Handler {
	return &Handler{Client: client}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/manifest", h.Manifest)
	r.POST("/invocations", h.Invoke)
	r.POST("/test", h.Test)
}

func (h *Handler) Manifest(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"protocol":   "provider/v1",
		"provider":   "synermaxx-sms-provider",
		"operations": []string{"send-sms"},
		"upstream":   "synermaxx/sendMessage",
	})
}

type invokeRequest struct {
	Operation string         `json:"operation"`
	Input     map[string]any `json:"input"`
}

func (h *Handler) Invoke(c *gin.Context) { h.handle(c) }
func (h *Handler) Test(c *gin.Context)   { h.handle(c) }

func (h *Handler) handle(c *gin.Context) {
	var req invokeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "detail": err.Error()})
		return
	}
	op := strings.TrimSpace(req.Operation)
	if op == "" {
		op = "send-sms"
	}
	if op != "send-sms" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported_operation", "detail": op})
		return
	}
	to, _ := req.Input["to"].(string)
	message, _ := req.Input["message"].(string)
	if strings.TrimSpace(to) == "" || strings.TrimSpace(message) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "invalid_input",
			"detail": "input.to and input.message are required",
		})
		return
	}
	corr, _ := c.Get("correlation_id")
	corrID, _ := corr.(string)

	mobile, err := synermaxx.NormalizePHMobile(to)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"outcome":        "rejected",
			"provider_ref":   "",
			"correlation_id": corrID,
			"detail":         gin.H{"error": err.Error(), "reason": "invalid_recipient"},
		})
		return
	}

	resp, status, err := h.Client.SendMessage(c.Request.Context(), synermaxx.SendMessageRequest{
		MobileNum:  mobile,
		Originator: h.Client.Originator,
		FullMesg:   message,
		Format:     "json",
		Blackhole:  "0",
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"outcome":        "provider_unavailable",
			"provider_ref":   "",
			"correlation_id": corrID,
			"detail":         gin.H{"error": err.Error()},
		})
		return
	}

	detail := resp.Raw
	if detail == nil {
		detail = map[string]any{"status": resp.Status, "http_status": status}
	}
	if synermaxx.IsAccepted(resp, status) {
		c.JSON(http.StatusOK, gin.H{
			"outcome":        "accepted",
			"provider_ref":   synermaxx.ProviderRef(resp),
			"correlation_id": corrID,
			"detail":         detail,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"outcome":        "rejected",
		"provider_ref":   synermaxx.ProviderRef(resp),
		"correlation_id": corrID,
		"detail":         detail,
	})
}
