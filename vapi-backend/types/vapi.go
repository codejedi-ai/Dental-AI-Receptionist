package types

import "encoding/json"

// Vapi webhook message types
const (
	TypeToolCalls          = "tool-calls"
	TypeStatusUpdate       = "status-update"
	TypeEndOfCallReport    = "end-of-call-report"
	TypeHang               = "hang"
	TypeSpeechUpdate       = "speech-update"
	TypeTranscript         = "transcript"
	TypeConversationUpdate = "conversation-update"
	TypeAssistantRequest   = "assistant-request"
)

// BaseVapiPayload is the common envelope for all Vapi webhook messages
type BaseVapiPayload struct {
	Type    string   `json:"type"`
	Call    VapiCall `json:"call"`
	CallID  string   `json:"callId,omitempty"`
}

// VapiCall represents the call object from Vapi webhooks
type VapiCall struct {
	ID            string `json:"id,omitempty"`
	AssistantID   string `json:"assistantId,omitempty"`
	PhoneNumberID string `json:"phoneNumberId,omitempty"`
	Status        string `json:"status,omitempty"`
	Type          string `json:"type,omitempty"`
}

// ToolCall represents a single tool call from Vapi
type ToolCall struct {
	ID        string          `json:"toolCallId"`
	ID2       string          `json:"id"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
	Function  struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	} `json:"function"`
}

// CallID returns the tool call ID, trying both field variants
func (t *ToolCall) CallID() string {
	if t.ID != "" {
		return t.ID
	}
	return t.ID2
}

// ToolName returns the tool name from top-level or nested function
func (t *ToolCall) ToolName() string {
	if t.Name != "" {
		return t.Name
	}
	return t.Function.Name
}

// Args returns the tool arguments, preferring top-level over nested
func (t *ToolCall) Args() json.RawMessage {
	if len(t.Arguments) > 0 {
		return t.Arguments
	}
	return t.Function.Arguments
}

// ToolResult is the response format Vapi expects
type ToolResult struct {
	ToolCallID string `json:"toolCallId"`
	Result     any    `json:"result"`
}

// ToolCallsResponse is the response to a tool-calls webhook
type ToolCallsResponse struct {
	Results []ToolResult `json:"results"`
}

// StatusUpdatePayload represents a status-update webhook
type StatusUpdatePayload struct {
	BaseVapiPayload
	Status struct {
		Status string `json:"status"`
	} `json:"status"`
}

// EndOfCallReportPayload represents an end-of-call-report webhook
type EndOfCallReportPayload struct {
	BaseVapiPayload
	EndedReason string  `json:"endedReason"`
	Summary     *string `json:"summary,omitempty"`
}

// HangPayload represents a hang webhook
type HangPayload struct {
	BaseVapiPayload
}

// SpeechUpdatePayload represents a speech-update webhook
type SpeechUpdatePayload struct {
	BaseVapiPayload
	Status string `json:"status"`
	Role   string `json:"role"`
}

// TranscriptPayload represents a transcript webhook
type TranscriptPayload struct {
	BaseVapiPayload
	Role       string `json:"role"`
	Transcript string `json:"transcript"`
}
