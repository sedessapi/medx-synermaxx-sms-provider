// Package synermaxx implements the Synermaxx sendMessage operation from
// openapi/synermaxx-sendmsg.openapi.yaml. No medx-base dependencies.
package synermaxx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const sendPath = "/api/sendmsg.php"

// SendMessageRequest matches components.schemas.SendMessageRequest.
type SendMessageRequest struct {
	MobileNum  string `json:"mobilenum"`
	Originator string `json:"originator"`
	FullMesg   string `json:"fullmesg"`
	Format     string `json:"format,omitempty"`
	Blackhole  string `json:"blackhole,omitempty"`
}

// SendMessageResponse matches components.schemas.SendMessageResponse.
type SendMessageResponse struct {
	Code           int    `json:"code"`
	Status         string `json:"status"`
	TransID        string `json:"transid"`
	ClientOutbox   string `json:"client_outboxnum"`
	MobileNum      string `json:"mobilenum"`
	Originator     string `json:"originator"`
	TxtParts       int    `json:"txtparts"`
	DLRFlag        string `json:"dlr_flag"`
	CostCtr        bool   `json:"cost_ctr"`
	CtrMobileNum   int    `json:"ctr_mobilenum"`
	ValidMobileNum int    `json:"valid_mobilenum"`
	Timestamp      string `json:"timestamp"`
	Message        string `json:"message"`
	Error          string `json:"error"`
	Raw            map[string]any `json:"-"`
}

type Client struct {
	BaseURL    string
	AppKey     string
	Originator string
	HTTP       *http.Client
}

func NewClient(baseURL, appKey, originator string) *Client {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if originator == "" {
		originator = "22566100"
	}
	return &Client{
		BaseURL:    baseURL,
		AppKey:     strings.TrimSpace(appKey),
		Originator: originator,
		HTTP:       &http.Client{Timeout: 30 * time.Second},
	}
}

// SendMessage implements operationId sendMessage.
func (c *Client) SendMessage(ctx context.Context, req SendMessageRequest) (SendMessageResponse, int, error) {
	if c.AppKey == "" {
		return SendMessageResponse{}, 0, fmt.Errorf("appkey not configured")
	}
	if req.Originator == "" {
		req.Originator = c.Originator
	}
	if req.Format == "" {
		req.Format = "json"
	}
	if req.Blackhole == "" {
		req.Blackhole = "0"
	}
	body, err := json.Marshal(req)
	if err != nil {
		return SendMessageResponse{}, 0, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+sendPath, bytes.NewReader(body))
	if err != nil {
		return SendMessageResponse{}, 0, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("appkey", c.AppKey)

	resp, err := c.HTTP.Do(httpReq)
	if err != nil {
		return SendMessageResponse{}, 0, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return SendMessageResponse{}, resp.StatusCode, err
	}

	var out SendMessageResponse
	_ = json.Unmarshal(raw, &out)
	var asMap map[string]any
	if json.Unmarshal(raw, &asMap) == nil {
		out.Raw = asMap
	}
	if out.Status == "" && len(raw) > 0 {
		out.Message = strings.TrimSpace(string(raw))
	}
	return out, resp.StatusCode, nil
}

// NormalizePHMobile maps E.164 / local PH numbers to 09XXXXXXXXX per OpenAPI pattern.
func NormalizePHMobile(recipient string) (string, error) {
	r := strings.TrimSpace(recipient)
	r = strings.ReplaceAll(r, " ", "")
	r = strings.ReplaceAll(r, "-", "")
	r = strings.TrimPrefix(r, "+")
	switch {
	case strings.HasPrefix(r, "63") && len(r) == 12:
		return "0" + r[2:], nil
	case strings.HasPrefix(r, "09") && len(r) == 11:
		return r, nil
	case strings.HasPrefix(r, "9") && len(r) == 10:
		return "0" + r, nil
	default:
		return "", fmt.Errorf("unsupported Philippine mobile format: %q", recipient)
	}
}

func IsAccepted(resp SendMessageResponse, httpStatus int) bool {
	if httpStatus < 200 || httpStatus >= 300 {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(resp.Status), "ACK")
}

func ProviderRef(resp SendMessageResponse) string {
	return strings.TrimRight(strings.TrimSpace(resp.TransID), ",")
}
