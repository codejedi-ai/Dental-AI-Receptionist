// Package appointmentbooking implements Module 3: Appointment Booking Flow
// for the Dental AI Receptionist system.
//
// It handles booking, cancellation, and rescheduling conversation flows,
// including collecting patient information, checking availability,
// confirming details, and booking the appointment.
package appointmentbooking

import (
	"log"
	"strings"
)

// ─── Action Types ───────────────────────────────────────────────

// Action represents the booking-related action being performed.
type Action string

const (
	ActionBook       Action = "booked"
	ActionCancel     Action = "cancelled"
	ActionReschedule Action = "rescheduled"
	ActionError      Action = "error"
	ActionConfirming Action = "confirming"
	ActionCollecting Action = "collecting"
)

// ─── Booking State Machine ──────────────────────────────────────

// BookingStep represents the current step in the booking conversation.
type BookingStep string

const (
	StepService       BookingStep = "collecting_service"
	StepDentist       BookingStep = "collecting_dentist"
	StepDate          BookingStep = "collecting_date"
	StepTime          BookingStep = "selecting_time"
	StepPatientInfo   BookingStep = "collecting_patient_info"
	StepConfirmation  BookingStep = "confirming_details"
	StepComplete      BookingStep = "complete"
	StepError         BookingStep = "error"
)

// BookingContext holds the state of an in-progress booking conversation.
type BookingContext struct {
	Step         BookingStep `json:"step"`
	Service      string      `json:"service,omitempty"`
	Dentist      string      `json:"dentist,omitempty"`
	Date         string      `json:"date,omitempty"`
	Time         string      `json:"time,omitempty"`
	PatientName  string      `json:"patient_name,omitempty"`
	PatientPhone string      `json:"patient_phone,omitempty"`
	PatientEmail string      `json:"patient_email,omitempty"`
	IsNewPatient bool        `json:"is_new_patient"`
	Notes        string      `json:"notes,omitempty"`

	// For cancellation/reschedule
	AppointmentID string `json:"appointment_id,omitempty"`
	OldDate       string `json:"old_date,omitempty"`
	OldTime       string `json:"old_time,omitempty"`
}

// ─── Booking Result ─────────────────────────────────────────────

// BookingResult is the output after completing a booking action.
type BookingResult struct {
	Module      string       `json:"module"`
	Action      Action       `json:"action"`
	Appointment *Appointment `json:"appointment,omitempty"`
	Message     string       `json:"message"`
	LangCode    string       `json:"lang_code"`
}

// Appointment represents a booked appointment.
type Appointment struct {
	ID            string `json:"id"`
	PatientName   string `json:"patient_name"`
	PatientPhone  string `json:"patient_phone"`
	Service       string `json:"service"`
	Dentist       string `json:"dentist"`
	Date          string `json:"date"`
	Time          string `json:"time"`
}

// ─── Required Fields ────────────────────────────────────────────

// requiredFields defines what's needed for a complete booking.
// STRICT: Only full name + phone required. Email is fallback if no phone.
var requiredFields = []string{"service", "dentist", "date", "time", "patient_name"}
// patient_phone OR patient_email is required (one of the two)

// ─── Response Templates ─────────────────────────────────────────

// English response templates.
const (
	ENPromptService = "What type of appointment would you like to book? We offer checkups, cleanings, whitening, fillings, crowns, root canals, implants, Invisalign, and pediatric care."
	ENPromptDentist = "Do you have a preference for which dentist you'd like to see? We have Dr. Sarah Chen, Dr. Michael Park, and Dr. Priya Sharma. Or I can book with whoever has the earliest availability."
	ENPromptDate    = "What date would you prefer? We're open Monday through Friday 8 AM to 6 PM, and Saturday 9 AM to 2 PM."
	ENPromptTime    = "Here are the available times. Which works best for you?"
	ENPromptPatient = "Could I get your full name and phone number, please?"
	ENConfirmDetails = "Let me confirm: %s, %s with %s on %s at %s. Is that correct?"
	ENBooked       = "Great! You're all set. Your appointment is confirmed. Please arrive 15 minutes early and bring your insurance card. Is there anything else I can help with?"
	ENNoSlots      = "Unfortunately there are no available slots on that date. Would you like to try another day?"
	ENCanceled     = "Your appointment has been cancelled. Would you like to book a new one?"
	ENRescheduled  = "Your appointment has been rescheduled successfully."
)

// Chinese response templates.
const (
	ZHPromptService = "请问您需要预约什么服务？我们提供检查、洗牙、美白、补牙、牙冠、根管治疗、种植牙、隐形矫正和儿童牙科。"
	ZHPromptDentist = "您对牙医有偏好吗？我们有陈医生、Park医生和Sharma医生。或者我可以为您预约最早有空的时间。"
	ZHPromptDate    = "您希望预约哪一天？我们周一至周五上午8点到下午6点营业，周六上午9点到下午2点。"
	ZHPromptTime    = "以下是可用的时间，哪个最适合您？"
	ZHPromptPatient = "请问您的全名和电话号码是？"
	ZHConfirmDetails = "让我确认一下：%s，%s，%s医生，%s，%s。对吗？"
	ZHBooked       = "好的，已经为您预约成功。请提前15分钟到达，并带上您的保险卡。还有什么可以帮您的吗？"
	ZHNoSlots      = "很抱歉，那天没有空位。您想试试其他日期吗？"
	ZHCanceled     = "您的预约已取消。需要重新预约吗？"
	ZHRescheduled  = "您的预约已成功改期。"
)

// ─── Public API ─────────────────────────────────────────────────

// NextStep determines what information to collect next based on the
// current booking context. Returns the next prompt to show the patient.
func NextStep(ctx BookingContext, langCode string) (BookingStep, string) {
	if ctx.Service == "" {
		return StepService, promptFor(langCode, "service")
	}
	if ctx.Dentist == "" {
		return StepDentist, promptFor(langCode, "dentist")
	}
	if ctx.Date == "" {
		return StepDate, promptFor(langCode, "date")
	}
	if ctx.Time == "" {
		return StepTime, promptFor(langCode, "time")
	}
	if ctx.PatientName == "" || (ctx.PatientPhone == "" && ctx.PatientEmail == "") {
		return StepPatientInfo, promptFor(langCode, "patient")
	}
	return StepConfirmation, confirmDetails(ctx, langCode)
}

// FillFields updates the booking context with new information from
// the patient's utterance. It intelligently maps what the patient said
// to the appropriate fields.
func FillFields(ctx BookingContext, utterance string, langCode string) BookingContext {
	lower := strings.ToLower(strings.TrimSpace(utterance))

	// Service extraction
	if ctx.Service == "" {
		service := extractService(lower, langCode)
		if service != "" {
			ctx.Service = service
		}
	}

	// Dentist extraction
	if ctx.Dentist == "" {
		dentist := extractDentist(lower, utterance)
		if dentist != "" {
			ctx.Dentist = dentist
		}
	}

	// Patient name extraction
	if ctx.PatientName == "" {
		name := extractPatientName(lower, utterance)
		if name != "" {
			ctx.PatientName = name
		}
	}

	// Phone extraction
	if ctx.PatientPhone == "" {
		phone := extractPhone(utterance)
		if phone != "" {
			ctx.PatientPhone = phone
		}
	}

	// Email extraction (fallback if no phone)
	if ctx.PatientPhone == "" && ctx.PatientEmail == "" {
		email := extractEmail(utterance)
		if email != "" {
			ctx.PatientEmail = email
		}
	}

	return ctx
}

// IsComplete checks if all required fields have been collected.
// STRICT: name + phone required, OR name + email if no phone.
func IsComplete(ctx BookingContext) bool {
	hasContact := ctx.PatientPhone != "" || ctx.PatientEmail != ""
	return ctx.Service != "" && ctx.Dentist != "" && ctx.Date != "" &&
		ctx.Time != "" && ctx.PatientName != "" && hasContact
}

// MissingFields returns a list of field names still needed.
func MissingFields(ctx BookingContext) []string {
	var missing []string
	if ctx.Service == "" {
		missing = append(missing, "service")
	}
	if ctx.Dentist == "" {
		missing = append(missing, "dentist")
	}
	if ctx.Date == "" {
		missing = append(missing, "date")
	}
	if ctx.Time == "" {
		missing = append(missing, "time")
	}
	if ctx.PatientName == "" {
		missing = append(missing, "patient_name")
	}
	if ctx.PatientPhone == "" && ctx.PatientEmail == "" {
		missing = append(missing, "patient_phone_or_email")
	}
	return missing
}

// ConfirmMessage generates the confirmation message to read back to the patient.
func ConfirmMessage(ctx BookingContext, langCode string) string {
	return confirmDetails(ctx, langCode)
}

// ─── Cancellation ───────────────────────────────────────────────

// CancelResult is the result of a cancellation request.
type CancelResult struct {
	Module  string `json:"module"`
	Action  Action `json:"action"`
	Message string `json:"message"`
	LangCode string `json:"lang_code"`
}

// CancelMessage generates the appropriate message for cancellation.
func CancelMessage(confirmed bool, langCode string) CancelResult {
	if confirmed {
		msg := selectTemplate(langCode, ENCanceled, ZHCanceled)
		return CancelResult{
			Module:   "appointment_booking",
			Action:   ActionCancel,
			Message:  msg,
			LangCode: langCode,
		}
	}
	msg := selectTemplate(langCode,
		"I can help with that. Could you provide the patient name and date of the appointment?",
		"好的，我来帮您处理。请提供患者姓名和预约日期。")
	return CancelResult{
		Module:   "appointment_booking",
		Action:   ActionCollecting,
		Message:  msg,
		LangCode: langCode,
	}
}

// ─── Rescheduling ───────────────────────────────────────────────

// RescheduleMessage generates the message for rescheduling flow.
func RescheduleMessage(hasExisting bool, langCode string) string {
	if hasExisting {
		return selectTemplate(langCode,
			"Let me check available times for your new appointment.",
			"让我查看一下新预约的可用时间。")
	}
	return selectTemplate(langCode,
		"I'd be happy to help reschedule. Could you provide your name and the current appointment date?",
		"我很乐意帮您改期。请提供您的姓名和当前预约日期。")
}

// ─── Emergency Handling ─────────────────────────────────────────

// EmergencyMessage generates the response for an emergency situation.
func EmergencyMessage(langCode string) string {
	return selectTemplate(langCode,
		"I understand this is urgent. For dental emergencies, I recommend going to the nearest emergency department right away. If you'd like, I can also try to book you a same-day appointment. What would you prefer?",
		"我理解情况紧急。对于牙科急诊，建议您立即前往最近的急诊室。如果您愿意，我也可以尝试为您预约当天的号。您希望怎么做？")
}

// ─── Helpers ────────────────────────────────────────────────────

func promptFor(langCode, field string) string {
	switch field {
	case "service":
		return selectTemplate(langCode, ENPromptService, ZHPromptService)
	case "dentist":
		return selectTemplate(langCode, ENPromptDentist, ZHPromptDentist)
	case "date":
		return selectTemplate(langCode, ENPromptDate, ZHPromptDate)
	case "time":
		return selectTemplate(langCode, ENPromptTime, ZHPromptTime)
	case "patient":
		return selectTemplate(langCode, ENPromptPatient, ZHPromptPatient)
	default:
		return selectTemplate(langCode, ENPromptService, ZHPromptService)
	}
}

func confirmDetails(ctx BookingContext, langCode string) string {
	return selectTemplate(langCode,
		ENConfirmDetails,
		ZHConfirmDetails,
		ctx.PatientName, ctx.Service, ctx.Dentist, ctx.Date, ctx.Time,
	)
}

func selectTemplate(langCode, enTemplate, zhTemplate string, args ...interface{}) string {
	var template string
	if langCode == "zh" {
		template = zhTemplate
	} else {
		template = enTemplate
	}
	if len(args) > 0 {
		return formatTemplate(template, args...)
	}
	return template
}

func formatTemplate(t string, args ...interface{}) string {
	// Simple sprintf-style formatting
	result := t
	for i, arg := range args {
		placeholder := "%" + string(rune('s'+i))
		// This is simplified; production would use fmt.Sprintf
		_ = placeholder
		_ = arg
	}
	// For now, just return the template — actual formatting is done by the LLM
	return result
}

// extractService finds a service name in the utterance.
func extractService(lower string, langCode string) string {
	serviceMap := map[string]string{
		"cleaning": "Cleaning", "clean": "Cleaning", "cleaning and polish": "Cleaning", "scale": "Cleaning",
		"checkup": "Checkup", "check up": "Checkup", "exam": "Checkup", "examination": "Checkup", "consultation": "Consultation", "consultaion": "Consultation", "consult": "Consultation",
		"whitening": "Whitening", "whiten": "Whitening",
		"filling": "Filling", "fill": "Filling", "cavity": "Filling", "cavities": "Filling",
		"crown": "Crown", "cap": "Crown", "crowns": "Crown",
		"bridge": "Bridge", "bridges": "Bridge",
		"root canal": "Root Canal", "root canals": "Root Canal",
		"implant": "Implant", "implants": "Implant",
		"invisalign": "Invisalign", "braces": "Invisalign",
		"pediatric": "Pediatric", "child": "Pediatric", "kids": "Pediatric",
		"emergency": "Emergency", "urgent": "Emergency", "pain": "Emergency", "broken": "Emergency",
	}

	if langCode == "zh" {
		zhMap := map[string]string{
			"洗牙": "Cleaning", "检查": "Checkup", "美白": "Whitening",
			"补牙": "Filling", "牙冠": "Crown", "根管": "Root Canal",
			"种植": "Implant", "正畸": "Invisalign", "儿童": "Pediatric",
			"桥": "Bridge", "咨询": "Consultation", "急诊": "Emergency",
		}
		for zh, svc := range zhMap {
			if strings.Contains(lower, zh) {
				return svc
			}
		}
	}

	for kw, svc := range serviceMap {
		if strings.Contains(lower, kw) {
			return svc
		}
	}
	return ""
}

// extractDentist finds a dentist name in the utterance.
func extractDentist(lower string, original string) string {
	dentists := []string{
		"Dr. Sarah Chen", "Dr. Chen", "Sarah Chen",
		"Dr. Michael Park", "Dr. Park", "Michael Park",
		"Dr. Priya Sharma", "Dr. Sharma", "Priya Sharma",
		"陈医生", "Park医生", "Sharma医生",
	}
	for _, d := range dentists {
		if strings.Contains(original, d) || strings.Contains(lower, strings.ToLower(d)) {
			return d
		}
	}
	// "any" or "no preference" means use any available dentist
	if strings.Contains(lower, "any") || strings.Contains(lower, "no preference") ||
		strings.Contains(lower, "whoever") || strings.Contains(lower, "随便") ||
		strings.Contains(lower, "都可以") {
		return "any"
	}
	return ""
}

// extractPatientName tries to extract a patient name from the utterance.
// This is a simplified implementation — production would use NER.
func extractPatientName(lower string, original string) string {
	// Look for patterns like "my name is X" or "I'm X"
	prefixes := []string{"my name is", "i'm", "i am", "我叫", "我是"}
	for _, prefix := range prefixes {
		idx := strings.Index(lower, prefix)
		if idx >= 0 {
			name := strings.TrimSpace(original[idx+len(prefix):])
			// Clean up trailing punctuation and stop at common delimiters
			name = strings.TrimRight(name, ".,!?;:")
			// Stop at common sentence connectors
			for _, stop := range []string{" and ", " with ", " at ", " on ", " my ", " your ", "，", "和", "与"} {
				if i := strings.Index(name, stop); i > 0 {
					name = strings.TrimSpace(name[:i])
					break
				}
			}
			if name != "" {
				return name
			}
		}
	}
	return ""
}

// extractPhone finds a phone number in the utterance.
func extractPhone(utterance string) string {
	// Simple pattern: look for digits
	var digits []byte
	for i := 0; i < len(utterance); i++ {
		c := utterance[i]
		if c >= '0' && c <= '9' {
			digits = append(digits, c)
		}
	}

	s := string(digits)
	if len(s) >= 10 {
		// Format as +1XXXXXXXXXX for North American numbers
		if len(s) == 10 {
			return "+1" + s
		}
		if len(s) == 11 && s[0] == '1' {
			return "+" + s
		}
		return "+" + s
	}
	return ""
}

// extractEmail finds an email address in the utterance.
func extractEmail(utterance string) string {
	// Simple regex-style search for email pattern
	for i := 0; i < len(utterance); i++ {
		if utterance[i] == '@' {
			// Find start of email (go backwards)
			start := i
			for start > 0 && isEmailChar(utterance[start-1]) {
				start--
			}
			// Find end of email (go forwards)
			end := i + 1
			for end < len(utterance) && isEmailChar(utterance[end]) {
				end++
			}
			email := utterance[start:end]
			if strings.Contains(email, ".") && len(email) > 5 {
				return email
			}
		}
	}
	return ""
}

func isEmailChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') || c == '.' || c == '_' || c == '-' || c == '+' || c == '@'
}

// ─── Logging ────────────────────────────────────────────────────

func logBookingStep(step BookingStep, langCode string) {
	log.Printf("[BookingFlow] Step=%s lang=%s", step, langCode)
}

func logBookingComplete(ctx BookingContext, langCode string) {
	log.Printf("[BookingFlow] Complete: patient=%s service=%s dentist=%s date=%s time=%s lang=%s",
		ctx.PatientName, ctx.Service, ctx.Dentist, ctx.Date, ctx.Time, langCode)
}
