package webhook

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"dental-ai-vapi/config"
	"dental-ai-vapi/types"

	"github.com/gin-gonic/gin"
)

// WebhookHandler handles all Vapi webhook messages at POST /api/tools
func WebhookHandler(cfg config.Config) gin.HandlerFunc {
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
			handleToolCalls(c, msgBody)
		case types.TypeStatusUpdate, types.TypeConversationUpdate:
			c.JSON(http.StatusOK, gin.H{})
		case types.TypeEndOfCallReport:
			var p types.EndOfCallReportPayload
			if err := json.Unmarshal(msgBody, &p); err == nil && p.Summary != nil {
				log.Printf("📋 End of call — summary: %s", *p.Summary)
			}
			c.JSON(http.StatusOK, gin.H{})
		case types.TypeHang, types.TypeSpeechUpdate:
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

func handleToolCalls(c *gin.Context, msg json.RawMessage) {
	// Parse all possible tool call array shapes
	var envelope struct {
		Message *struct {
			Type         string          `json:"type"`
			ToolCalls    []types.ToolCall `json:"tool_calls"`
			ToolCallsCC  []types.ToolCall `json:"toolCalls"`
			ToolCallList []types.ToolCall `json:"toolCallList"`
		} `json:"message"`
		Type         string          `json:"type"`
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

		result, err := dispatchTool(&call)
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

func dispatchTool(call *types.ToolCall) (any, error) {
	name := call.ToolName()
	args := call.Args()

	log.Printf("  → %s args=%s", name, string(args))

	switch name {
	case "get_dentists":
		return getDentists()
	case "get_current_date":
		return getCurrentDate(), nil
	case "get_clinic_info":
		return getClinicInfo(args)
	case "check_availability":
		return checkAvailability(args)
	case "parse_date":
		return parseDate(args)
	case "get_next_available_dates":
		return getNextAvailableDates(args)
	case "book_appointment":
		return bookAppointment(args)
	case "cancel_appointment":
		return cancelAppointment(args)
	case "send_booking_confirmation":
		return sendBookingConfirmation(args)
	case "validate_patient_info":
		return validatePatientInfo(args)
	case "is_booking_complete":
		return isBookingComplete(args)
	case "get_booking_step":
		return getBookingStep(args)
	case "fill_booking_fields":
		return fillBookingFields(args)
	case "get_confirm_message":
		return getConfirmMessage(args)
	case "get_cancel_message":
		return getCancelMessage(args)
	case "get_reschedule_message":
		return getRescheduleMessage(args)
	case "get_emergency_message":
		return getEmergencyMessage(args)
	case "detect_language":
		return detectLanguage(args)
	case "classify_intent":
		return classifyIntent(args)
	case "transfer_to_chinese_agent":
		return "Transferring to Li, the Chinese-speaking agent.", nil
	case "transfer_to_english_agent":
		return "Transferring to Riley, the English-speaking agent.", nil
	default:
		return map[string]any{
			"error":   true,
			"message": fmt.Sprintf("unknown tool: %s", name),
		}, nil
	}
}

// Stub implementations — these will call into internal/db and internal/tools
// In the next step, wire these to your existing implementations

func getDentists() (string, error) {
	// TODO: wire to db.GetDentists()
	return "Our dentists are: Dr. Michael Park, Dr. Priya Sharma, Dr. Sarah Chen.", nil
}

func getCurrentDate() string {
	// TODO: wire to tools.GetCurrentDate()
	return "Today is Thursday, April 9, 2026."
}

func getClinicInfo(args json.RawMessage) (string, error) {
	// TODO: wire to tools.GetClinicInfo()
	return "Clinic info: Smile Dental Clinic, open Mon-Fri 9am-5pm.", nil
}

func checkAvailability(args json.RawMessage) (string, error) {
	// TODO: wire to tools.CheckAvailability()
	return "Available slots found.", nil
}

func parseDate(args json.RawMessage) (string, error) {
	// TODO: wire to tools.ParseDate()
	return "2026-04-10", nil
}

func getNextAvailableDates(args json.RawMessage) (string, error) {
	// TODO: wire to tools.GetNextAvailableDates()
	return "Next available: 2026-04-10, 2026-04-13, 2026-04-14.", nil
}

func bookAppointment(args json.RawMessage) (string, error) {
	// TODO: wire to tools.BookAppointment()
	return "Appointment booked successfully.", nil
}

func cancelAppointment(args json.RawMessage) (string, error) {
	// TODO: wire to tools.CancelAppointment()
	return "Appointment cancelled.", nil
}

func sendBookingConfirmation(args json.RawMessage) (string, error) {
	// TODO: wire to tools.SendBookingConfirmation()
	return "Confirmation sent.", nil
}

func validatePatientInfo(args json.RawMessage) (string, error) {
	// TODO: wire to tools.ValidatePatientInfo()
	return `{"valid": true}`, nil
}

func isBookingComplete(args json.RawMessage) (string, error) {
	// TODO: wire to tools.IsBookingComplete()
	return `{"complete": false, "missing": []}`, nil
}

func getBookingStep(args json.RawMessage) (string, error) {
	// TODO: wire to tools.GetBookingStep()
	return `{"step": "collect_name", "message": "What is your full name?"}`, nil
}

func fillBookingFields(args json.RawMessage) (string, error) {
	// TODO: wire to tools.FillBookingFields()
	return `{}`, nil
}

func getConfirmMessage(args json.RawMessage) (string, error) {
	// TODO: wire to tools.GetConfirmMessage()
	return `{"message": "Please confirm your appointment."}`, nil
}

func getCancelMessage(args json.RawMessage) (string, error) {
	// TODO: wire to tools.GetCancelMessage()
	return `{"message": "Your appointment has been cancelled."}`, nil
}

func getRescheduleMessage(args json.RawMessage) (string, error) {
	// TODO: wire to tools.GetRescheduleMessage()
	return `{"message": "Let's reschedule your appointment."}`, nil
}

func getEmergencyMessage(args json.RawMessage) (string, error) {
	// TODO: wire to tools.GetEmergencyMessage()
	return `{"message": "For emergencies, please call 911 or visit the nearest emergency room."}`, nil
}

func detectLanguage(args json.RawMessage) (string, error) {
	// TODO: wire to languagedetection.Detect()
	return `{"lang_code": "en", "confidence": 0.9}`, nil
}

func classifyIntent(args json.RawMessage) (string, error) {
	// TODO: wire to intentclassifier.Classify()
	return `{"intent": "booking", "confidence": 0.9}`, nil
}
