package appointmentbooking

import (
	"testing"
)

// ─── NextStep Tests ─────────────────────────────────────────────

func TestNextStep_ServiceFirst(t *testing.T) {
	ctx := BookingContext{}
	step, msg := NextStep(ctx, "en")
	if step != StepService {
		t.Errorf("expected step=collecting_service, got %s", step)
	}
	if msg != ENPromptService {
		t.Errorf("expected service prompt, got %s", msg)
	}
}

func TestNextStep_ServiceChinese(t *testing.T) {
	ctx := BookingContext{}
	_, msg := NextStep(ctx, "zh")
	if msg != ZHPromptService {
		t.Errorf("expected Chinese service prompt, got %s", msg)
	}
}

func TestNextStep_DentistAfterService(t *testing.T) {
	ctx := BookingContext{Service: "Cleaning"}
	step, msg := NextStep(ctx, "en")
	if step != StepDentist {
		t.Errorf("expected step=collecting_dentist, got %s", step)
	}
	if msg != ENPromptDentist {
		t.Errorf("expected dentist prompt, got %s", msg)
	}
}

func TestNextStep_DateAfterDentist(t *testing.T) {
	ctx := BookingContext{Service: "Cleaning", Dentist: "Dr. Sarah Chen"}
	step, msg := NextStep(ctx, "en")
	if step != StepDate {
		t.Errorf("expected step=collecting_date, got %s", step)
	}
	if msg != ENPromptDate {
		t.Errorf("expected date prompt, got %s", msg)
	}
}

func TestNextStep_TimeAfterDate(t *testing.T) {
	ctx := BookingContext{Service: "Cleaning", Dentist: "Dr. Sarah Chen", Date: "2026-04-15"}
	step, msg := NextStep(ctx, "en")
	if step != StepTime {
		t.Errorf("expected step=selecting_time, got %s", step)
	}
	if msg != ENPromptTime {
		t.Errorf("expected time prompt, got %s", msg)
	}
}

func TestNextStep_PatientInfoAfterTime(t *testing.T) {
	ctx := BookingContext{Service: "Cleaning", Dentist: "Dr. Sarah Chen", Date: "2026-04-15", Time: "10:00"}
	step, msg := NextStep(ctx, "en")
	if step != StepPatientInfo {
		t.Errorf("expected step=collecting_patient_info, got %s", step)
	}
	if msg != ENPromptPatient {
		t.Errorf("expected patient info prompt, got %s", msg)
	}
}

func TestNextStep_ConfirmationWhenComplete(t *testing.T) {
	ctx := BookingContext{
		Service:      "Cleaning",
		Dentist:      "Dr. Sarah Chen",
		Date:         "2026-04-15",
		Time:         "10:00",
		PatientName:  "John Doe",
		PatientPhone: "+1234567890",
	}
	step, _ := NextStep(ctx, "en")
	if step != StepConfirmation {
		t.Errorf("expected step=confirming_details, got %s", step)
	}
}

// ─── FillFields Tests ───────────────────────────────────────────

func TestFillFields_ServiceExtraction(t *testing.T) {
	ctx := FillFields(BookingContext{}, "I'd like a cleaning please", "en")
	if ctx.Service != "Cleaning" {
		t.Errorf("expected service=Cleaning, got %s", ctx.Service)
	}
}

func TestFillFields_ServiceExtractionChinese(t *testing.T) {
	ctx := FillFields(BookingContext{}, "我想洗牙", "zh")
	if ctx.Service != "Cleaning" {
		t.Errorf("expected service=Cleaning for Chinese, got %s", ctx.Service)
	}
}

func TestFillFields_DentistExtraction(t *testing.T) {
	ctx := FillFields(BookingContext{}, "I'd like to see Dr. Sarah Chen", "en")
	if ctx.Dentist != "Dr. Sarah Chen" {
		t.Errorf("expected dentist=Dr. Sarah Chen, got %s", ctx.Dentist)
	}
}

func TestFillFields_DentistAnyPreference(t *testing.T) {
	ctx := FillFields(BookingContext{}, "I don't have a preference, anyone is fine", "en")
	if ctx.Dentist != "any" {
		t.Errorf("expected dentist=any, got %s", ctx.Dentist)
	}
}

func TestFillFields_DentistAnyPreferenceChinese(t *testing.T) {
	ctx := FillFields(BookingContext{}, "随便哪位医生都可以", "zh")
	if ctx.Dentist != "any" {
		t.Errorf("expected dentist=any for Chinese, got %s", ctx.Dentist)
	}
}

func TestFillFields_PatientNameExtraction(t *testing.T) {
	ctx := FillFields(BookingContext{}, "My name is John Smith", "en")
	if ctx.PatientName != "John Smith" {
		t.Errorf("expected patient_name=John Smith, got %s", ctx.PatientName)
	}
}

func TestFillFields_PatientNameIm(t *testing.T) {
	ctx := FillFields(BookingContext{}, "I'm Jane Doe", "en")
	if ctx.PatientName != "Jane Doe" {
		t.Errorf("expected patient_name=Jane Doe, got %s", ctx.PatientName)
	}
}

func TestFillFields_PatientNameChinese(t *testing.T) {
	ctx := FillFields(BookingContext{}, "我叫张三", "zh")
	if ctx.PatientName != "张三" {
		t.Errorf("expected patient_name=张三 for Chinese, got %s", ctx.PatientName)
	}
}

func TestFillFields_PhoneExtraction(t *testing.T) {
	ctx := FillFields(BookingContext{}, "My number is 4165551234", "en")
	if ctx.PatientPhone != "+14165551234" {
		t.Errorf("expected patient_phone=+14165551234, got %s", ctx.PatientPhone)
	}
}

func TestFillFields_PhoneExtraction11Digits(t *testing.T) {
	ctx := FillFields(BookingContext{}, "Call me at 14165551234", "en")
	if ctx.PatientPhone != "+14165551234" {
		t.Errorf("expected patient_phone=+14165551234, got %s", ctx.PatientPhone)
	}
}

func TestFillFields_MultipleFields(t *testing.T) {
	ctx := FillFields(BookingContext{}, "I need a cleaning with Dr. Chen. My name is Bob and my phone is 4165559999", "en")
	if ctx.Service != "Cleaning" {
		t.Errorf("expected service=Cleaning, got %s", ctx.Service)
	}
	// Dr. Chen should be matched
	if ctx.Dentist != "Dr. Chen" {
		t.Errorf("expected dentist=Dr. Chen, got %s", ctx.Dentist)
	}
	if ctx.PatientName != "Bob" {
		t.Errorf("expected patient_name=Bob, got %s", ctx.PatientName)
	}
	if ctx.PatientPhone != "+14165559999" {
		t.Errorf("expected patient_phone=+14165559999, got %s", ctx.PatientPhone)
	}
}

func TestFillFields_NoOverwrite(t *testing.T) {
	ctx := BookingContext{Service: "Whitening"}
	ctx = FillFields(ctx, "I need a cleaning", "en")
	// Should not overwrite existing service
	if ctx.Service != "Whitening" {
		t.Errorf("expected service to remain Whitening, got %s", ctx.Service)
	}
}

// ─── IsComplete Tests ───────────────────────────────────────────

func TestIsComplete_AllFields(t *testing.T) {
	ctx := BookingContext{
		Service:      "Cleaning",
		Dentist:      "Dr. Sarah Chen",
		Date:         "2026-04-15",
		Time:         "10:00",
		PatientName:  "John Doe",
		PatientPhone: "+1234567890",
	}
	if !IsComplete(ctx) {
		t.Error("expected complete with all fields")
	}
}

func TestIsComplete_MissingService(t *testing.T) {
	ctx := BookingContext{
		Dentist:      "Dr. Sarah Chen",
		Date:         "2026-04-15",
		Time:         "10:00",
		PatientName:  "John Doe",
		PatientPhone: "+1234567890",
	}
	if IsComplete(ctx) {
		t.Error("expected incomplete without service")
	}
}

func TestIsComplete_MissingPhone(t *testing.T) {
	ctx := BookingContext{
		Service:      "Cleaning",
		Dentist:      "Dr. Sarah Chen",
		Date:         "2026-04-15",
		Time:         "10:00",
		PatientName:  "John Doe",
	}
	if IsComplete(ctx) {
		t.Error("expected incomplete without phone")
	}
}

// ─── MissingFields Tests ────────────────────────────────────────

func TestMissingFields_Empty(t *testing.T) {
	ctx := BookingContext{}
	missing := MissingFields(ctx)
	if len(missing) != 6 {
		t.Errorf("expected 6 missing fields, got %d", len(missing))
	}
}

func TestMissingFields_Partial(t *testing.T) {
	ctx := BookingContext{Service: "Cleaning", Dentist: "Dr. Chen"}
	missing := MissingFields(ctx)
	if len(missing) != 4 {
		t.Errorf("expected 4 missing fields, got %d: %v", len(missing), missing)
	}
}

func TestMissingFields_None(t *testing.T) {
	ctx := BookingContext{
		Service:      "Cleaning",
		Dentist:      "Dr. Sarah Chen",
		Date:         "2026-04-15",
		Time:         "10:00",
		PatientName:  "John Doe",
		PatientPhone: "+1234567890",
	}
	missing := MissingFields(ctx)
	if len(missing) != 0 {
		t.Errorf("expected 0 missing fields, got %d: %v", len(missing), missing)
	}
}

// ─── ConfirmMessage Tests ───────────────────────────────────────

func TestConfirmMessage_English(t *testing.T) {
	ctx := BookingContext{
		PatientName: "John Doe",
		Service:     "Cleaning",
		Dentist:     "Dr. Sarah Chen",
		Date:        "2026-04-15",
		Time:        "10:00",
	}
	msg := ConfirmMessage(ctx, "en")
	if msg != ENConfirmDetails {
		t.Errorf("expected English confirm template, got %s", msg)
	}
}

func TestConfirmMessage_Chinese(t *testing.T) {
	ctx := BookingContext{
		PatientName: "张三",
		Service:     "洗牙",
		Dentist:     "陈医生",
		Date:        "2026-04-15",
		Time:        "10:00",
	}
	msg := ConfirmMessage(ctx, "zh")
	if msg != ZHConfirmDetails {
		t.Errorf("expected Chinese confirm template, got %s", msg)
	}
}

// ─── Cancellation Tests ─────────────────────────────────────────

func TestCancelMessage_Confirmed(t *testing.T) {
	r := CancelMessage(true, "en")
	if r.Action != ActionCancel {
		t.Errorf("expected action=cancelled, got %s", r.Action)
	}
	if r.Message != ENCanceled {
		t.Errorf("expected cancel message, got %s", r.Message)
	}
}

func TestCancelMessage_Collecting(t *testing.T) {
	r := CancelMessage(false, "en")
	if r.Action != ActionCollecting {
		t.Errorf("expected action=collecting, got %s", r.Action)
	}
}

func TestCancelMessage_Chinese(t *testing.T) {
	r := CancelMessage(true, "zh")
	if r.Message != ZHCanceled {
		t.Errorf("expected Chinese cancel message, got %s", r.Message)
	}
}

// ─── Rescheduling Tests ─────────────────────────────────────────

func TestRescheduleMessage_HasExisting(t *testing.T) {
	msg := RescheduleMessage(true, "en")
	if msg != "Let me check available times for your new appointment." {
		t.Errorf("unexpected reschedule message: %s", msg)
	}
}

func TestRescheduleMessage_NoExisting(t *testing.T) {
	msg := RescheduleMessage(false, "en")
	if msg != "I'd be happy to help reschedule. Could you provide your name and the current appointment date?" {
		t.Errorf("unexpected reschedule message: %s", msg)
	}
}

func TestRescheduleMessage_Chinese(t *testing.T) {
	msg := RescheduleMessage(true, "zh")
	if msg != "让我查看一下新预约的可用时间。" {
		t.Errorf("unexpected Chinese reschedule message: %s", msg)
	}
}

// ─── Emergency Tests ────────────────────────────────────────────

func TestEmergencyMessage_English(t *testing.T) {
	msg := EmergencyMessage("en")
	if msg != "I understand this is urgent. For dental emergencies, I recommend going to the nearest emergency department right away. If you'd like, I can also try to book you a same-day appointment. What would you prefer?" {
		t.Errorf("unexpected emergency message: %s", msg)
	}
}

func TestEmergencyMessage_Chinese(t *testing.T) {
	msg := EmergencyMessage("zh")
	if msg != "我理解情况紧急。对于牙科急诊，建议您立即前往最近的急诊室。如果您愿意，我也可以尝试为您预约当天的号。您希望怎么做？" {
		t.Errorf("unexpected Chinese emergency message: %s", msg)
	}
}

// ─── Service Extraction Tests ───────────────────────────────────

func TestExtractService_Cleaning(t *testing.T) {
	s := extractService("i need a cleaning", "en")
	if s != "Cleaning" {
		t.Errorf("expected Cleaning, got %s", s)
	}
}

func TestExtractService_RootCanal(t *testing.T) {
	s := extractService("i need a root canal", "en")
	if s != "Root Canal" {
		t.Errorf("expected Root Canal, got %s", s)
	}
}

func TestExtractService_Chinese(t *testing.T) {
	s := extractService("我想洗牙", "zh")
	if s != "Cleaning" {
		t.Errorf("expected Cleaning for 洗牙, got %s", s)
	}
}

// ─── Phone Extraction Tests ─────────────────────────────────────

func TestExtractPhone_10Digits(t *testing.T) {
	p := extractPhone("4165551234")
	if p != "+14165551234" {
		t.Errorf("expected +14165551234, got %s", p)
	}
}

func TestExtractPhone_11Digits(t *testing.T) {
	p := extractPhone("14165551234")
	if p != "+14165551234" {
		t.Errorf("expected +14165551234, got %s", p)
	}
}

func TestExtractPhone_InSentence(t *testing.T) {
	p := extractPhone("My number is 416-555-1234")
	if p != "+14165551234" {
		t.Errorf("expected +14165551234, got %s", p)
	}
}

func TestExtractPhone_TooShort(t *testing.T) {
	p := extractPhone("12345")
	if p != "" {
		t.Errorf("expected empty for short number, got %s", p)
	}
}

// ─── Booking Context Tests ──────────────────────────────────────

func TestBookingContext_DefaultValues(t *testing.T) {
	ctx := BookingContext{}
	if ctx.Step != "" {
		t.Error("expected default step to be empty")
	}
	if ctx.IsNewPatient != false {
		t.Error("expected default is_new_patient to be false")
	}
}

// ─── Booking Result Tests ───────────────────────────────────────

func TestBookingResult_Module(t *testing.T) {
	r := BookingResult{Module: "appointment_booking"}
	if r.Module != "appointment_booking" {
		t.Errorf("expected module=appointment_booking, got %s", r.Module)
	}
}
