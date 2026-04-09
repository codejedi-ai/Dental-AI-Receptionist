package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"dental-ai-vapi/internal/db"
	"dental-ai-vapi/internal/tools"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Handler struct {
	pg         *db.Postgres
	mongo      *db.Mongo
	webhookURL string
}

func New(pg *db.Postgres, mongo *db.Mongo, webhookURL string) *Handler {
	return &Handler{pg: pg, mongo: mongo, webhookURL: webhookURL}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "dental-ai-vapi"})
}

func (h *Handler) Verify(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		h.verifyPage(w, "Invalid link", "This verification link is invalid.", false)
		return
	}
	ok := h.mongo.MarkVerified(r.Context(), token)
	if ok {
		h.verifyPage(w, "Email Verified!", "Your email has been verified. You're all set — your appointment will be confirmed shortly.", true)
	} else {
		h.verifyPage(w, "Link Expired", "This verification link has expired or was already used. Please contact the clinic if you need to rebook.", false)
	}
}

func (h *Handler) verifyPage(w http.ResponseWriter, title, body string, success bool) {
	color := "#c0392b"
	icon := "⚠️"
	if success {
		color = "#1a6fc4"
		icon = "✅"
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html><html><head><meta charset="utf-8"><title>%s</title>
<meta name="viewport" content="width=device-width,initial-scale=1">
<style>body{font-family:sans-serif;display:flex;align-items:center;justify-content:center;min-height:100vh;margin:0;background:#f4f6f9}
.card{background:#fff;border-radius:12px;padding:40px 32px;max-width:420px;text-align:center;box-shadow:0 4px 20px rgba(0,0,0,.1)}
h1{color:%s;margin-top:8px}p{color:#555;line-height:1.6}.icon{font-size:48px}</style></head>
<body><div class="card"><div class="icon">%s</div><h1>%s</h1><p>%s</p>
<p style="margin-top:32px;font-size:13px;color:#999">Dental Clinic</p></div></body></html>`,
		title, color, icon, title, body)
}

func (h *Handler) Tools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Message json.RawMessage `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var msgType struct {
		Type string `json:"type"`
	}
	json.Unmarshal(req.Message, &msgType)

	switch msgType.Type {
	case "tool-calls":
		h.handleToolCalls(w, req.Message)
	case "status-update", "conversation-update", "end-of-call-report", "speech-update", "hang":
		h.logStatusUpdate(req.Message)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{})
	default:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{})
	}
}

type ToolCall struct {
	ID       string `json:"id"`
	Function struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	} `json:"function"`
}

type ToolResult struct {
	ToolCallID string `json:"toolCallId"`
	Result     string `json:"result"`
}

func (h *Handler) handleToolCalls(w http.ResponseWriter, msg json.RawMessage) {
	var tc struct {
		ToolCalls []ToolCall `json:"tool_calls"`
	}
	if err := json.Unmarshal(msg, &tc); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"results": []string{}})
		return
	}

	ctx := context.Background()
	results := make([]ToolResult, 0, len(tc.ToolCalls))

	for _, call := range tc.ToolCalls {
		log.Printf("🔧 Tool call: %s", call.Function.Name)
		var result string
		var status string

		switch call.Function.Name {
		case "check_availability":
			result, status = tools.CheckAvailability(ctx, h.pg, call.Function.Arguments)
		case "book_appointment":
			result, status = tools.BookAppointment(ctx, h.pg, h.mongo, call.Function.Arguments)
		case "cancel_appointment":
			result, status = tools.CancelAppointment(ctx, h.pg, h.mongo, call.Function.Arguments)
		case "get_clinic_info":
			result, status = tools.GetClinicInfo(ctx, h.mongo, call.Function.Arguments)
		case "send_sms_code":
			result, status = tools.SendSMSCode(call.Function.Arguments)
		case "verify_sms_code":
			result, status = tools.VerifySMSCode(call.Function.Arguments)
		case "lookup_patient":
			result, status = tools.LookupPatient(ctx, h.pg, call.Function.Arguments)
		case "send_booking_confirmation":
			result, status = tools.SendBookingConfirmation(call.Function.Arguments)
		case "get_dentists":
			result, status = tools.GetDentists(ctx, h.pg)
		case "get_current_date":
			result, status = tools.GetCurrentDate()
		default:
			result = fmt.Sprintf("Unknown tool: %s", call.Function.Name)
			status = "error"
		}

		// Audit log
		var args primitive.M
		json.Unmarshal(call.Function.Arguments, &args)
		h.mongo.LogToolCall(ctx, call.Function.Name, args, result, status)

		results = append(results, ToolResult{ToolCallID: call.ID, Result: result})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"results": results})
}

func (h *Handler) logStatusUpdate(msg json.RawMessage) {
	var s struct {
		Type   string `json:"type"`
		Status struct {
			Status string `json:"status"`
		} `json:"status"`
		Summary        *string `json:"summary"`
		DurationSeconds *int64 `json:"durationSeconds"`
	}
	json.Unmarshal(msg, &s)

	if s.Type == "end-of-call-report" && s.Summary != nil {
		log.Printf("📋 End of call — summary: %s", *s.Summary)
		h.mongo.LogCall(context.Background(), nil, s.DurationSeconds, "ended", *s.Summary)
	} else {
		log.Printf("📊 Status: %s", s.Status.Status)
	}
}

// Helper for tools

func ParseArg(args json.RawMessage, key string) string {
	var m map[string]interface{}
	json.Unmarshal(args, &m)
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func ParseArgOpt(args json.RawMessage, key string) *string {
	var m map[string]interface{}
	json.Unmarshal(args, &m)
	if v, ok := m[key].(string); ok && v != "" {
		return &v
	}
	return nil
}

func TimeString(t time.Time) string {
	h := t.Hour()
	m := t.Minute()
	period := "AM"
	if h >= 12 {
		period = "PM"
	}
	if h > 12 {
		h -= 12
	}
	if h == 0 {
		h = 12
	}
	return fmt.Sprintf("%d:%02d %s", h, m, period)
}
