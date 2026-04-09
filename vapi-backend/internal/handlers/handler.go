package handlers

import (
	"crypto/subtle"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"dental-ai-vapi/internal/db"
	"dental-ai-vapi/internal/modules/appointmentbooking"
	"dental-ai-vapi/internal/modules/intentclassifier"
	"dental-ai-vapi/internal/modules/languagedetection"
	"dental-ai-vapi/internal/tools"
	"dental-ai-vapi/internal/util"
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
	// GET allows quick reachability checks (browser / cellular) without a POST body.
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":      true,
			"service": "dental-ai-vapi-tools",
			"hint":    "Vapi sends POST with JSON body (type tool-calls); include Authorization: Bearer <TOOL_API_KEY>",
		})
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, errRead := io.ReadAll(r.Body)
	if errRead != nil {
		log.Printf("⚠️  POST /api/tools read body: %v", errRead)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	logVapiToolsIngress(r, body, "incoming")

	if !isAuthorizedToolRequest(r) {
		if vapiToolsDebugEnabled() {
			log.Printf("[vapi-tools-debug] rejected 401 — compare TOOL_API_KEY with Vapi tool server Bearer or ?token=")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized tool request"})
		return
	}

	log.Printf("📨 POST /api/tools body: %.500s", string(body))

	// Vapi sends type either at top level or inside "message"
	var topLevel struct {
		Type    string `json:"type"`
		Message json.RawMessage `json:"message"`
	}
	json.Unmarshal(body, &topLevel)

	msgType := topLevel.Type
	var msgBody json.RawMessage

	if msgType == "" && len(topLevel.Message) > 0 {
		// Shape: { "message": { "type": "tool-calls", ... } }
		var inner struct {
			Type string `json:"type"`
		}
		json.Unmarshal(topLevel.Message, &inner)
		msgType = inner.Type
		msgBody = topLevel.Message
	} else if msgType != "" {
		// Shape: { "type": "tool-calls", "toolCalls": [...] } (top-level)
		msgBody = body
	}

	switch msgType {
	case "tool-calls":
		h.handleToolCalls(w, msgBody)
	case "status-update", "conversation-update", "end-of-call-report", "speech-update", "hang":
		h.logStatusUpdate(msgBody)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{})
	default:
		log.Printf("⚠️  Unknown message type: %q (body: %.200s)", msgType, string(body))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{})
	}
}

type ToolCall struct {
	ID              string          `json:"toolCallId"`    // variant 1
	ID2             string          `json:"id"`            // variant 2: Vapi also sends just "id"
	Name            string          `json:"name"`          // direct name field
	Arguments       json.RawMessage `json:"arguments"`     // direct arguments
	Function        struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	} `json:"function"`
}

// CallID returns the tool call ID, trying both field variants.
func (t *ToolCall) CallID() string {
	if t.ID != "" {
		return t.ID
	}
	return t.ID2
}

// Args returns the tool arguments, preferring top-level over nested.
func (t *ToolCall) Args() json.RawMessage {
	if len(t.Arguments) > 0 {
		return t.Arguments
	}
	return t.Function.Arguments
}

// ToolName returns the tool name from top-level or nested function.
func (t *ToolCall) ToolName() string {
	if t.Name != "" {
		return t.Name
	}
	return t.Function.Name
}

type ToolResult struct {
	ToolCallID string `json:"toolCallId"`
	Result     any    `json:"result"`
}

func vapiToolsDebugEnabled() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv("VAPI_TOOLS_DEBUG")))
	return v == "1" || v == "true" || v == "yes"
}

func authHeaderDebugSummary(auth string) string {
	auth = strings.TrimSpace(auth)
	if auth == "" {
		return "absent"
	}
	la := strings.ToLower(auth)
	if strings.HasPrefix(la, "bearer ") {
		tok := strings.TrimSpace(auth[7:])
		if tok == "" {
			return "Bearer (empty token)"
		}
		n := len(tok)
		tail := tok
		if n > 6 {
			tail = tok[n-6:]
		}
		return fmt.Sprintf("Bearer redacted len=%d …%s", n, tail)
	}
	return fmt.Sprintf("non-Bearer len=%d", len(auth))
}

// logVapiToolsIngress logs URL, selected headers (Authorization redacted), and body when VAPI_TOOLS_DEBUG is set.
func logVapiToolsIngress(r *http.Request, body []byte, phase string) {
	if !vapiToolsDebugEnabled() {
		return
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if xf := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); xf != "" {
		scheme = strings.ToLower(xf)
	}
	log.Printf("[vapi-tools-debug] phase=%s | %s %s://%s%s | RemoteAddr=%s",
		phase, r.Method, scheme, r.Host, r.URL.RequestURI(), r.RemoteAddr)
	log.Printf("[vapi-tools-debug]   UA=%q | Content-Type=%q | Content-Length=%q | Authorization: %s",
		r.UserAgent(), r.Header.Get("Content-Type"), r.Header.Get("Content-Length"), authHeaderDebugSummary(r.Header.Get("Authorization")))
	for _, k := range []string{"X-Forwarded-For", "X-Real-Ip", "Cf-Connecting-Ip", "X-Vapi-Signature"} {
		if v := r.Header.Get(k); v != "" {
			log.Printf("[vapi-tools-debug]   %s=%q", k, v)
		}
	}
	if len(body) == 0 {
		log.Printf("[vapi-tools-debug]   body: <empty>")
		return
	}
	const maxBody = 16384
	s := string(body)
	if len(s) > maxBody {
		log.Printf("[vapi-tools-debug]   body (%d bytes, showing first %d): %s … [truncated]",
			len(body), maxBody, s[:maxBody])
		return
	}
	log.Printf("[vapi-tools-debug]   body (%d bytes): %s", len(body), s)
}

func logVapiToolsResponse(payload any, elapsed time.Duration) {
	if !vapiToolsDebugEnabled() {
		return
	}
	b, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[vapi-tools-debug] response marshal error after %s: %v", elapsed, err)
		return
	}
	s := string(b)
	const max = 8192
	if len(s) > max {
		log.Printf("[vapi-tools-debug] response after %s (%d bytes, truncated): %s …", elapsed, len(s), s[:max])
		return
	}
	log.Printf("[vapi-tools-debug] response after %s: %s", elapsed, s)
}

func (h *Handler) handleToolCalls(w http.ResponseWriter, msg json.RawMessage) {
	start := time.Now()
	log.Printf("📨 Raw tool message: %s", string(msg))

	// Vapi sends tool calls in multiple possible shapes.
	// Try them all in order of likelihood based on live logs.
	var envelope struct {
		// Shape A: wrapped in message
		Message *struct {
			Type         string     `json:"type"`
			ToolCalls    []ToolCall `json:"tool_calls"`
			ToolCallsCC  []ToolCall `json:"toolCalls"`     // camelCase (live Vapi)
			ToolCallList []ToolCall `json:"toolCallList"`  // alternative
		} `json:"message"`

		// Shape B: top-level (no message wrapper)
		Type         string     `json:"type"`
		ToolCalls    []ToolCall `json:"tool_calls"`
		ToolCallsCC  []ToolCall `json:"toolCalls"`
		ToolCallList []ToolCall `json:"toolCallList"`
	}

	if err := json.Unmarshal(msg, &envelope); err != nil {
		log.Printf("⚠️  JSON unmarshal error: %v", err)
		payload := map[string]any{"results": []string{}}
		logVapiToolsResponse(payload, time.Since(start))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
		return
	}

	// Extract calls: check message-wrapped first, then top-level
	var calls []ToolCall
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
		payload := map[string]any{"results": []string{}}
		logVapiToolsResponse(payload, time.Since(start))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
		return
	}

	ctx := context.Background()
	results := make([]ToolResult, 0, len(calls))

	for _, call := range calls {
		log.Printf("🔧 Tool call: name=%s, callId=%s", call.ToolName(), call.CallID())
		var result string
		var status string

		switch call.ToolName() {
		case "check_availability":
			result, status = tools.CheckAvailability(ctx, h.pg, call.Args())
		case "book_appointment":
			result, status = tools.BookAppointment(ctx, h.pg, h.mongo, call.Args())
		case "cancel_appointment":
			result, status = tools.CancelAppointment(ctx, h.pg, h.mongo, call.Args())
		case "get_clinic_info":
			result, status = tools.GetClinicInfo(ctx, h.mongo, call.Args())
		case "lookup_patient":
			result, status = tools.LookupPatient(ctx, h.pg, call.Args())
		case "send_booking_confirmation":
			result, status = tools.SendBookingConfirmation(call.Args())
		case "get_dentists":
			result, status = tools.GetDentists(ctx, h.pg)
		case "get_current_date":
			result, status = tools.GetCurrentDate()
		case "validate_patient_info":
			result, status = tools.ValidatePatientInfo(call.Args())
		case "parse_date":
			result, status = tools.ParseDate(ctx, call.Args())
		case "get_next_available_dates":
			result, status = tools.GetNextAvailableDates(ctx, h.pg, call.Args())
		case "detect_language":
			result, status = handleDetectLanguage(call.Args())
		case "classify_intent":
			result, status = handleClassifyIntent(call.Args())
		case "get_booking_step":
			result, status = handleGetBookingStep(call.Args())
		case "fill_booking_fields":
			result, status = handleFillBookingFields(call.Args())
		case "is_booking_complete":
			result, status = handleIsBookingComplete(call.Args())
		case "get_confirm_message":
			result, status = handleGetConfirmMessage(call.Args())
		case "get_cancel_message":
			result, status = handleGetCancelMessage(call.Args())
		case "get_reschedule_message":
			result, status = handleGetRescheduleMessage(call.Args())
		case "get_emergency_message":
			result, status = handleGetEmergencyMessage(call.Args())
		case "transfer_to_chinese_agent":
			result = "Transferring to Li, the Chinese-speaking agent."
			status = "success"
		case "transfer_to_english_agent":
			result = "Transferring to Riley, the English-speaking agent."
			status = "success"
		default:
			result = fmt.Sprintf("Unknown tool: %s", call.ToolName())
			status = "error"
		}

		// Audit log → file (logs/tool_calls_YYYY-MM-DD.log)
		util.LogToolCall(call.ToolName(), call.Args(), result, status)

		results = append(results, ToolResult{ToolCallID: call.CallID(), Result: result})
	}

	payload := map[string]interface{}{"results": results}
	logVapiToolsResponse(payload, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
}

func isAuthorizedToolRequest(r *http.Request) bool {
	expected := strings.TrimSpace(os.Getenv("TOOL_API_KEY"))
	if expected == "" {
		// Also accept query-param token if set by Vapi (serverUrl?token=...)
		queryToken := strings.TrimSpace(r.URL.Query().Get("token"))
		if queryToken != "" {
			return true
		}
		return true
	}
	// Prefer Authorization header
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		provided := strings.TrimSpace(auth[len("Bearer "):])
		if provided != "" && subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1 {
			return true
		}
	}
	// Fall back to query param
	queryToken := strings.TrimSpace(r.URL.Query().Get("token"))
	if queryToken != "" && subtle.ConstantTimeCompare([]byte(queryToken), []byte(expected)) == 1 {
		return true
	}
	return false
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

// ─── Module 1: Language Detection ───────────────────────────────

func handleDetectLanguage(args json.RawMessage) (string, string) {
	sentence := ParseArg(args, "sentence")
	if sentence == "" {
		return `{"error":"sentence is required"}`, "error"
	}
	result := languagedetection.Detect(sentence)
	data, _ := json.Marshal(result)
	return string(data), "success"
}

// ─── Module 2: Intent Classification ────────────────────────────

func handleClassifyIntent(args json.RawMessage) (string, string) {
	utterance := ParseArg(args, "utterance")
	langCode := ParseArg(args, "lang_code")
	if utterance == "" {
		return `{"error":"utterance is required"}`, "error"
	}
	if langCode == "" {
		langCode = "en"
	}
	result := intentclassifier.Classify(utterance, langCode)
	data, _ := json.Marshal(result)
	return string(data), "success"
}

// ─── Module 3: Appointment Booking ──────────────────────────────

func handleGetBookingStep(args json.RawMessage) (string, string) {
	var ctx appointmentbooking.BookingContext
	if err := json.Unmarshal(args, &ctx); err != nil {
		return `{"error":"invalid booking context"}`, "error"
	}
	langCode := ParseArg(args, "lang_code")
	if langCode == "" {
		langCode = "en"
	}
	step, msg := appointmentbooking.NextStep(ctx, langCode)
	data, _ := json.Marshal(map[string]interface{}{
		"step":    step,
		"message": msg,
	})
	return string(data), "success"
}

func handleFillBookingFields(args json.RawMessage) (string, string) {
	var ctx appointmentbooking.BookingContext
	if err := json.Unmarshal(args, &ctx); err != nil {
		return `{"error":"invalid booking context"}`, "error"
	}
	utterance := ParseArg(args, "utterance")
	langCode := ParseArg(args, "lang_code")
	if langCode == "" {
		langCode = "en"
	}
	// Also try to extract email from the raw args if provided directly
	if ctx.PatientEmail == "" {
		ctx.PatientEmail = ParseArg(args, "patient_email")
	}
	updated := appointmentbooking.FillFields(ctx, utterance, langCode)
	data, _ := json.Marshal(updated)
	return string(data), "success"
}

func handleIsBookingComplete(args json.RawMessage) (string, string) {
	var ctx appointmentbooking.BookingContext
	if err := json.Unmarshal(args, &ctx); err != nil {
		return `{"error":"invalid booking context"}`, "error"
	}
	complete := appointmentbooking.IsComplete(ctx)
	missing := appointmentbooking.MissingFields(ctx)
	data, _ := json.Marshal(map[string]interface{}{
		"complete": complete,
		"missing":  missing,
	})
	return string(data), "success"
}

func handleGetConfirmMessage(args json.RawMessage) (string, string) {
	var ctx appointmentbooking.BookingContext
	if err := json.Unmarshal(args, &ctx); err != nil {
		return `{"error":"invalid booking context"}`, "error"
	}
	langCode := ParseArg(args, "lang_code")
	if langCode == "" {
		langCode = "en"
	}
	msg := appointmentbooking.ConfirmMessage(ctx, langCode)
	data, _ := json.Marshal(map[string]string{"message": msg})
	return string(data), "success"
}

func handleGetCancelMessage(args json.RawMessage) (string, string) {
	confirmed := ParseArg(args, "confirmed") == "true"
	langCode := ParseArg(args, "lang_code")
	if langCode == "" {
		langCode = "en"
	}
	result := appointmentbooking.CancelMessage(confirmed, langCode)
	data, _ := json.Marshal(result)
	return string(data), "success"
}

func handleGetRescheduleMessage(args json.RawMessage) (string, string) {
	hasExisting := ParseArg(args, "has_existing") == "true"
	langCode := ParseArg(args, "lang_code")
	if langCode == "" {
		langCode = "en"
	}
	msg := appointmentbooking.RescheduleMessage(hasExisting, langCode)
	data, _ := json.Marshal(map[string]string{"message": msg})
	return string(data), "success"
}

func handleGetEmergencyMessage(args json.RawMessage) (string, string) {
	langCode := ParseArg(args, "lang_code")
	if langCode == "" {
		langCode = "en"
	}
	msg := appointmentbooking.EmergencyMessage(langCode)
	data, _ := json.Marshal(map[string]string{"message": msg})
	return string(data), "success"
}
