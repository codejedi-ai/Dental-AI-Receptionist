package intentclassifier

import (
	"testing"
)

// ─── Booking Intent Tests ───────────────────────────────────────

func TestClassify_BookAppointment(t *testing.T) {
	r := Classify("I'd like to book a cleaning", "en")
	if r.Intent != IntentBook {
		t.Errorf("expected intent=book_appointment, got %s", r.Intent)
	}
	if r.Confidence < ConfidenceThreshold {
		t.Errorf("expected confidence >= %.2f, got %.2f", ConfidenceThreshold, r.Confidence)
	}
	if r.NextModule != "appointment_booking" {
		t.Errorf("expected next_module=appointment_booking, got %s", r.NextModule)
	}
}

func TestClassify_BookAppointmentSchedule(t *testing.T) {
	r := Classify("I want to schedule an appointment", "en")
	if r.Intent != IntentBook {
		t.Errorf("expected intent=book_appointment, got %s", r.Intent)
	}
}

func TestClassify_BookAppointmentChinese(t *testing.T) {
	r := Classify("我想预约", "zh")
	if r.Intent != IntentBook {
		t.Errorf("expected intent=book_appointment for Chinese, got %s", r.Intent)
	}
	if r.Confidence < ConfidenceThreshold {
		t.Errorf("expected confidence >= %.2f for Chinese booking, got %.2f", ConfidenceThreshold, r.Confidence)
	}
}

// ─── Cancellation Intent Tests ─────────────────────────────────

func TestClassify_CancelAppointment(t *testing.T) {
	r := Classify("I need to cancel my appointment", "en")
	if r.Intent != IntentCancel {
		t.Errorf("expected intent=cancel_appointment, got %s", r.Intent)
	}
	if r.NextModule != "appointment_cancellation" {
		t.Errorf("expected next_module=appointment_cancellation, got %s", r.NextModule)
	}
}

func TestClassify_CancelAppointmentChinese(t *testing.T) {
	r := Classify("我想取消我的预约", "zh")
	if r.Intent != IntentCancel {
		t.Errorf("expected intent=cancel_appointment for Chinese, got %s", r.Intent)
	}
}

// ─── Reschedule Intent Tests ────────────────────────────────────

func TestClassify_RescheduleAppointment(t *testing.T) {
	r := Classify("Can I reschedule my appointment to Friday?", "en")
	if r.Intent != IntentReschedule {
		t.Errorf("expected intent=reschedule_appointment, got %s", r.Intent)
	}
	if r.NextModule != "appointment_reschedule" {
		t.Errorf("expected next_module=appointment_reschedule, got %s", r.NextModule)
	}
}

func TestClassify_RescheduleChangeTime(t *testing.T) {
	r := Classify("I need to change my appointment time", "en")
	if r.Intent != IntentReschedule {
		t.Errorf("expected intent=reschedule_appointment, got %s", r.Intent)
	}
}

func TestClassify_RescheduleChinese(t *testing.T) {
	r := Classify("我想改期", "zh")
	if r.Intent != IntentReschedule {
		t.Errorf("expected intent=reschedule_appointment for Chinese, got %s", r.Intent)
	}
}

// ─── Hours Inquiry Tests ────────────────────────────────────────

func TestClassify_AskHours(t *testing.T) {
	r := Classify("What are your hours?", "en")
	if r.Intent != IntentAskHours {
		t.Errorf("expected intent=ask_hours, got %s", r.Intent)
	}
	if r.NextModule != "knowledge_base" {
		t.Errorf("expected next_module=knowledge_base, got %s", r.NextModule)
	}
}

func TestClassify_AskHoursWhenOpen(t *testing.T) {
	r := Classify("When are you open?", "en")
	if r.Intent != IntentAskHours {
		t.Errorf("expected intent=ask_hours for 'when are you open', got %s", r.Intent)
	}
}

func TestClassify_AskHoursChinese(t *testing.T) {
	r := Classify("你们几点开门？", "zh")
	if r.Intent != IntentAskHours {
		t.Errorf("expected intent=ask_hours for Chinese, got %s", r.Intent)
	}
}

// ─── Location Inquiry Tests ─────────────────────────────────────

func TestClassify_AskLocation(t *testing.T) {
	r := Classify("Where are you located?", "en")
	if r.Intent != IntentAskLocation {
		t.Errorf("expected intent=ask_location, got %s", r.Intent)
	}
	if r.NextModule != "knowledge_base" {
		t.Errorf("expected next_module=knowledge_base, got %s", r.NextModule)
	}
}

func TestClassify_AskLocationAddress(t *testing.T) {
	r := Classify("What's your address?", "en")
	if r.Intent != IntentAskLocation {
		t.Errorf("expected intent=ask_location for address query, got %s", r.Intent)
	}
}

func TestClassify_AskLocationChinese(t *testing.T) {
	r := Classify("你们在哪里？", "zh")
	if r.Intent != IntentAskLocation {
		t.Errorf("expected intent=ask_location for Chinese, got %s", r.Intent)
	}
}

// ─── Services Inquiry Tests ─────────────────────────────────────

func TestClassify_AskServices(t *testing.T) {
	r := Classify("Do you do root canals?", "en")
	if r.Intent != IntentAskServices {
		t.Errorf("expected intent=ask_services, got %s", r.Intent)
	}
}

func TestClassify_AskServicesGeneral(t *testing.T) {
	r := Classify("What services do you offer?", "en")
	if r.Intent != IntentAskServices {
		t.Errorf("expected intent=ask_services, got %s", r.Intent)
	}
}

func TestClassify_AskServicesChinese(t *testing.T) {
	r := Classify("你们提供什么服务？", "zh")
	if r.Intent != IntentAskServices {
		t.Errorf("expected intent=ask_services for Chinese, got %s", r.Intent)
	}
}

// ─── Emergency Intent Tests ─────────────────────────────────────

func TestClassify_Emergency(t *testing.T) {
	r := Classify("I have a toothache, it's urgent", "en")
	if r.Intent != IntentEmergency {
		t.Errorf("expected intent=emergency, got %s", r.Intent)
	}
	if r.NextModule != "emergency_handler" {
		t.Errorf("expected next_module=emergency_handler, got %s", r.NextModule)
	}
	if r.Confidence != ConfidenceExactMatch {
		t.Errorf("expected confidence=exact_match (%.2f), got %.2f", ConfidenceExactMatch, r.Confidence)
	}
}

func TestClassify_EmergencyBrokenTooth(t *testing.T) {
	r := Classify("I chipped my tooth", "en")
	if r.Intent != IntentEmergency {
		t.Errorf("expected intent=emergency for broken tooth, got %s", r.Intent)
	}
}

func TestClassify_EmergencyChinese(t *testing.T) {
	r := Classify("我牙痛，很紧急", "zh")
	if r.Intent != IntentEmergency {
		t.Errorf("expected intent=emergency for Chinese, got %s", r.Intent)
	}
}

// ─── Greeting Intent Tests ──────────────────────────────────────

func TestClassify_Greeting(t *testing.T) {
	r := Classify("Hello", "en")
	if r.Intent != IntentGreeting {
		t.Errorf("expected intent=greeting, got %s", r.Intent)
	}
	if r.NextModule != "greeting_handler" {
		t.Errorf("expected next_module=greeting_handler, got %s", r.NextModule)
	}
}

func TestClassify_GreetingChinese(t *testing.T) {
	r := Classify("你好", "zh")
	if r.Intent != IntentGreeting {
		t.Errorf("expected intent=greeting for Chinese, got %s", r.Intent)
	}
}

// ─── Unclear Intent Tests ───────────────────────────────────────

func TestClassify_Unclear(t *testing.T) {
	r := Classify("ummm maybe I don't know", "en")
	if r.Intent != IntentUnclear {
		t.Errorf("expected intent=unclear, got %s", r.Intent)
	}
	if r.NextModule != "clarification_needed" {
		t.Errorf("expected next_module=clarification_needed, got %s", r.NextModule)
	}
}

func TestClassify_Empty(t *testing.T) {
	r := Classify("", "en")
	if r.Intent != IntentUnclear {
		t.Errorf("expected intent=unclear for empty input, got %s", r.Intent)
	}
	if r.Confidence != ConfidenceBelowThreshold {
		t.Errorf("expected confidence=below_threshold, got %.2f", r.Confidence)
	}
}

// ─── Entity Extraction Tests ────────────────────────────────────

func TestClassify_EntityServiceExtraction(t *testing.T) {
	r := Classify("I'd like to book a cleaning", "en")
	if r.Entities.Service == nil || *r.Entities.Service != "Cleaning" {
		t.Errorf("expected service=Cleaning, got %v", r.Entities.Service)
	}
}

func TestClassify_EntityServiceExtractionRootCanal(t *testing.T) {
	r := Classify("Do you do root canals?", "en")
	if r.Entities.Service == nil || *r.Entities.Service != "Root Canal" {
		t.Errorf("expected service=Root Canal, got %v", r.Entities.Service)
	}
}

func TestClassify_EntityServiceExtractionChinese(t *testing.T) {
	r := Classify("我想洗牙", "zh")
	if r.Entities.Service == nil || *r.Entities.Service != "Cleaning" {
		t.Errorf("expected service=Cleaning for Chinese 洗牙, got %v", r.Entities.Service)
	}
}

func TestClassify_EntityDateExtraction(t *testing.T) {
	r := Classify("I want to book for tomorrow", "en")
	if r.Entities.PreferredDate == nil || *r.Entities.PreferredDate != "tomorrow" {
		t.Errorf("expected preferred_date=tomorrow, got %v", r.Entities.PreferredDate)
	}
}

func TestClassify_EntityDateExtractionChinese(t *testing.T) {
	r := Classify("我想预约明天", "zh")
	if r.Entities.PreferredDate == nil || *r.Entities.PreferredDate != "tomorrow" {
		t.Errorf("expected preferred_date=tomorrow for Chinese, got %v", r.Entities.PreferredDate)
	}
}

func TestClassify_EntityTimeExtraction(t *testing.T) {
	r := Classify("I prefer the afternoon", "en")
	if r.Entities.PreferredTime == nil || *r.Entities.PreferredTime != "afternoon" {
		t.Errorf("expected preferred_time=afternoon, got %v", r.Entities.PreferredTime)
	}
}

func TestClassify_EntityDentistExtraction(t *testing.T) {
	r := Classify("I'd like to see Dr. Sarah Chen", "en")
	if r.Entities.Dentist == nil || *r.Entities.Dentist != "Dr. Sarah Chen" {
		t.Errorf("expected dentist=Dr. Sarah Chen, got %v", r.Entities.Dentist)
	}
}

func TestClassify_EntityDentistExtractionChinese(t *testing.T) {
	r := Classify("我想找陈医生", "zh")
	if r.Entities.Dentist == nil {
		t.Error("expected dentist to be extracted for Chinese")
	}
}

func TestClassify_EntityReasonExtraction(t *testing.T) {
	r := Classify("I have a toothache", "en")
	if r.Entities.Reason == nil {
		t.Error("expected reason to be extracted for toothache")
	}
}

func TestClassify_NoEntitiesForGreeting(t *testing.T) {
	r := Classify("Hello", "en")
	if r.Entities.Service != nil {
		t.Error("expected no service entity for greeting")
	}
	if r.Entities.PreferredDate != nil {
		t.Error("expected no date entity for greeting")
	}
}

// ─── Metadata Tests ─────────────────────────────────────────────

func TestClassify_ModuleField(t *testing.T) {
	r := Classify("Hello", "en")
	if r.Module != "intent_classification" {
		t.Errorf("expected module=intent_classification, got %s", r.Module)
	}
}

func TestClassify_LangCodePreserved(t *testing.T) {
	r := Classify("你好", "zh")
	if r.LangCode != "zh" {
		t.Errorf("expected lang_code=zh, got %s", r.LangCode)
	}
}

func TestClassify_MatchedRules(t *testing.T) {
	r := Classify("I'd like to book a cleaning", "en")
	if len(r.Metadata.MatchedRules) == 0 {
		t.Error("expected at least one matched rule")
	}
}

func TestClassify_InputSentencePreserved(t *testing.T) {
	input := "I'd like to book a cleaning"
	r := Classify(input, "en")
	if r.Metadata.InputSentence != input {
		t.Errorf("expected input_sentence=%q, got %q", input, r.Metadata.InputSentence)
	}
}

// ─── Confidence Ordering Tests ──────────────────────────────────

func TestClassify_EmergencyHigherThanBooking(t *testing.T) {
	// Emergency should score higher due to exact_match confidence
	emergency := Classify("I have a severe toothache emergency", "en")
	booking := Classify("I'd like to book an appointment", "en")
	if emergency.Confidence <= booking.Confidence {
		t.Errorf("expected emergency confidence (%.2f) > booking confidence (%.2f)",
			emergency.Confidence, booking.Confidence)
	}
}

// ─── Service Keyword Mapping Tests ──────────────────────────────

func TestExtractService_Cleaning(t *testing.T) {
	if s := extractService("i need a cleaning", "en"); s != "Cleaning" {
		t.Errorf("expected Cleaning, got %s", s)
	}
}

func TestExtractService_Whitening(t *testing.T) {
	if s := extractService("teeth whitening please", "en"); s != "Whitening" {
		t.Errorf("expected Whitening, got %s", s)
	}
}

func TestExtractService_Filling(t *testing.T) {
	if s := extractService("i have a cavity", "en"); s != "Filling" {
		t.Errorf("expected Filling, got %s", s)
	}
}

func TestExtractService_Implant(t *testing.T) {
	if s := extractService("i need an implant", "en"); s != "Implant" {
		t.Errorf("expected Implant, got %s", s)
	}
}

func TestExtractService_Invisalign(t *testing.T) {
	if s := extractService("do you do invisalign", "en"); s != "Invisalign" {
		t.Errorf("expected Invisalign, got %s", s)
	}
}

func TestExtractService_NoService(t *testing.T) {
	if s := extractService("hello how are you", "en"); s != "" {
		t.Errorf("expected no service, got %s", s)
	}
}

// ─── Date/Time Extraction Tests ─────────────────────────────────

func TestExtractDate_DayOfWeek(t *testing.T) {
	if d := extractDate("next monday"); d != "next_monday" {
		t.Errorf("expected next_monday, got %s", d)
	}
}

func TestExtractDate_Tomorrow(t *testing.T) {
	if d := extractDate("tomorrow"); d != "tomorrow" {
		t.Errorf("expected tomorrow, got %s", d)
	}
}

func TestExtractTime_Morning(t *testing.T) {
	if d := extractTime("in the morning"); d != "morning" {
		t.Errorf("expected morning, got %s", d)
	}
}

func TestExtractTime_NoTime(t *testing.T) {
	if d := extractTime("any time is fine"); d != "" {
		t.Errorf("expected no time, got %s", d)
	}
}
