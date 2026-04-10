package webhook

import (
	"crypto/subtle"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"dental-ai-vapi/config"
	"dental-ai-vapi/service"
	"dental-ai-vapi/types"

	"github.com/gin-gonic/gin"
)

// WebhookHandler handles all Vapi webhook messages at POST /api/tools
func WebhookHandler(cfg config.Config, dispatcher *service.ToolDispatcher) gin.HandlerFunc {
	return func(c *gin.Context) {
		// GET allows reachability checks
		if c.Request.Method == http.MethodGet {
			c.JSON(http.StatusOK, gin.H{
				"ok":      true,
				"service": "dental-ai-vapi-tools",
				"hint":    "Vapi sends POST with JSON body (type tool-calls); include Authorization: Bearer <TOOL_API_KEY>",
			})
			return
		}

		if c.Request.Method != http.MethodPost {
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method not allowed"})
			return
		}

		// Auth check
		if !isAuthorized(c, cfg.ToolAPIKey) {
			if cfg.ToolsDebug {
				log.Printf("[vapi-tools-debug] rejected 401")
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized tool request"})
			return
		}

		// Read body
		body, err := c.GetRawData()
		if err != nil {
			log.Printf("⚠️  POST /api/tools read body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
			return
		}

		if cfg.ToolsDebug {
			log.Printf("📨 POST /api/tools body: %.500s", string(body))
		}

		// Parse message type
		var raw struct {
			Type    string          `json:"type"`
			Message json.RawMessage `json:"message"`
		}
		json.Unmarshal(body, &raw)

		msgType := raw.Type
		var msgBody json.RawMessage

		if msgType == "" && len(raw.Message) > 0 {
			var inner struct {
				Type string `json:"type"`
			}
			json.Unmarshal(raw.Message, &inner)
			msgType = inner.Type
			msgBody = raw.Message
		} else if msgType != "" {
			msgBody = body
		}

		switch msgType {
		case types.TypeToolCalls:
			handleToolCalls(c, msgBody, dispatcher)
		case types.TypeStatusUpdate, types.TypeConversationUpdate, types.TypeEndOfCallReport, types.TypeHang, types.TypeSpeechUpdate, types.TypeTranscript:
			if msgType == types.TypeEndOfCallReport {
				var p types.EndOfCallReportPayload
				if err := json.Unmarshal(msgBody, &p); err == nil && p.Summary != nil {
					log.Printf("📋 End of call — summary: %s", *p.Summary)
				}
			}
			c.JSON(http.StatusOK, gin.H{})
		default:
			if cfg.ToolsDebug {
				log.Printf("⚠️  Unknown message type: %q", msgType)
			}
			c.JSON(http.StatusOK, gin.H{})
		}
	}
}

func isAuthorized(c *gin.Context, expected string) bool {
	expected = strings.TrimSpace(expected)
	if expected == "" {
		return true
	}

	// Check Authorization header
	auth := strings.TrimSpace(c.GetHeader("Authorization"))
	if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		token := strings.TrimSpace(auth[7:])
		if token != "" && subtle.ConstantTimeCompare([]byte(token), []byte(expected)) == 1 {
			return true
		}
	}

	// Check query param token
	if token := strings.TrimSpace(c.Query("token")); token != "" {
		if subtle.ConstantTimeCompare([]byte(token), []byte(expected)) == 1 {
			return true
		}
	}

	return false
}

func handleToolCalls(c *gin.Context, msg json.RawMessage, dispatcher *service.ToolDispatcher) {
	// Parse all possible tool call array shapes
	var envelope struct {
		Message *struct {
			Type         string           `json:"type"`
			ToolCalls    []types.ToolCall `json:"tool_calls"`
			ToolCallsCC  []types.ToolCall `json:"toolCalls"`
			ToolCallList []types.ToolCall `json:"toolCallList"`
		} `json:"message"`
		Type         string           `json:"type"`
		ToolCalls    []types.ToolCall `json:"tool_calls"`
		ToolCallsCC  []types.ToolCall `json:"toolCalls"`
		ToolCallList []types.ToolCall `json:"toolCallList"`
	}

	if err := json.Unmarshal(msg, &envelope); err != nil {
		log.Printf("⚠️  JSON unmarshal error: %v", err)
		c.JSON(http.StatusOK, gin.H{"results": []string{}})
		return
	}

	// Extract calls from any shape
	var calls []types.ToolCall
	if envelope.Message != nil {
		calls = envelope.Message.ToolCalls
		if len(calls) == 0 {
			calls = envelope.Message.ToolCallsCC
		}
		if len(calls) == 0 {
			calls = envelope.Message.ToolCallList
		}
	}
	if len(calls) == 0 {
		calls = envelope.ToolCalls
	}
	if len(calls) == 0 {
		calls = envelope.ToolCallsCC
	}
	if len(calls) == 0 {
		calls = envelope.ToolCallList
	}
	if len(calls) == 0 {
		log.Printf("⚠️  No tool calls found in payload")
		c.JSON(http.StatusOK, gin.H{"results": []string{}})
		return
	}

	results := make([]types.ToolResult, 0, len(calls))

	for _, call := range calls {
		log.Printf("🔧 Tool call: name=%s, callId=%s", call.ToolName(), call.CallID())

		result, err := dispatcher.Dispatch(call.ToolName(), call.Args())
		if err != nil {
			log.Printf("❌ Tool %s error: %v", call.ToolName(), err)
			result = map[string]any{"error": true, "message": err.Error()}
		}

		results = append(results, types.ToolResult{
			ToolCallID: call.CallID(),
			Result:     result,
		})
	}

	c.JSON(http.StatusOK, types.ToolCallsResponse{Results: results})
}
