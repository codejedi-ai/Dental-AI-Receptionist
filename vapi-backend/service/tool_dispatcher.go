package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
	_ "time/tzdata"

	"dental-ai-vapi/db"
)

type ToolDispatcher struct {
	pg    *db.Postgres
	mongo *db.Mongo
}

func NewToolDispatcher(pg *db.Postgres, mongo *db.Mongo) *ToolDispatcher {
	return &ToolDispatcher{pg: pg, mongo: mongo}
}

func (d *ToolDispatcher) Dispatch(name string, args json.RawMessage) (any, error) {
	ctx := context.Background()
	log.Printf("  → %s args=%s", name, string(args))

	switch name {
	case "get_dentists":
		return d.getDentists(ctx)
	case "get_current_date":
		return getCurrentDate(), nil
	case "get_clinic_info":
		return d.getClinicInfo(ctx, args)
	case "check_availability":
		return d.checkAvailability(ctx, args)
	case "parse_date":
		return parseDate(args)
	case "get_next_available_dates":
		return d.getNextAvailableDates(ctx, args)
	case "book_appointment":
		return d.bookAppointment(ctx, args)
	case "cancel_appointment":
		return d.cancelAppointment(ctx, args)
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

// ─── get_dentists ──────────────────────────────────────────────

func (d *ToolDispatcher) getDentists(ctx context.Context) (string, error) {
	names, err := d.pg.GetAllDentists(ctx)
	if err != nil || len(names) == 0 {
		log.Printf("[getDentists] error or empty: %v", err)
		return "I'm having trouble retrieving the dentist list right now.", nil
	}
	return fmt.Sprintf("Our dentists are: %s.", strings.Join(names, ", ")), nil
}

// ─── get_current_date ──────────────────────────────────────────

func getCurrentDate() string {
	now := time.Now()
	loc, _ := time.LoadLocation("America/Toronto")
	now = now.In(loc)
	date := now.Format("Monday, January 2, 2006")
	timeStr := now.Format("3:04 PM")
	return fmt.Sprintf("Today is %s, %s Toronto time.", date, timeStr)
}

// ─── get_clinic_info ───────────────────────────────────────────

func (d *ToolDispatcher) getClinicInfo(ctx context.Context, args json.RawMessage) (string, error) {
	var a struct {
		Topic string `json:"topic"`
	}
	json.Unmarshal(args, &a)
	if a.Topic == "" {
		a.Topic = "general"
	}
	return d.mongo.GetClinicInfo(ctx, a.Topic), nil
}

// ─── check_availability ─────────────────────────────────────────

var weekdaySlots = []string{
	"08:00", "08:30", "09:00", "09:30", "10:00", "10:30", "11:00", "11:30",
	"12:00", "12:30", "13:00", "13:30", "14:00", "14:30", "15:00", "15:30",
	"16:00", "16:30", "17:00", "17:30",
}
var saturdaySlots = []string{
	"09:00", "09:30", "10:00", "10:30", "11:00", "11:30",
	"12:00", "12:30", "13:00", "13:30",
}

func (d *ToolDispatcher) checkAvailability(ctx context.Context, args json.RawMessage) (string, error) {
	var a struct {
		Date    string `json:"date"`
		Dentist string `json:"dentist"`
		Service string `json:"service"`
	}
	json.Unmarshal(args, &a)

	if a.Date == "" {
		// fallback key
		a.Date = parseArg(args, "appointmentDate")
	}
	if a.Dentist == "any" {
		a.Dentist = ""
	}

	if strings.TrimSpace(a.Date) == "" {
		return "I need a date in YYYY-MM-DD format to check availability.", nil
	}

	requestedDate, err := time.Parse("2006-01-02", a.Date)
	if err != nil {
		return fmt.Sprintf("Invalid date format: %s. Please use YYYY-MM-DD.", a.Date), nil
	}

	today := time.Now().Truncate(24 * time.Hour)
	if requestedDate.Before(today) {
		return "That date is in the past. Please provide a future date.", nil
	}

	dow := requestedDate.Weekday()
	if dow == time.Sunday {
		return "Sorry, the clinic is closed on Sundays. We're open Monday–Friday 8 AM–6 PM and Saturday 9 AM–2 PM.", nil
	}

	slots := weekdaySlots
	if dow == time.Saturday {
		slots = saturdaySlots
	}

	allDentists, err := d.pg.GetAllDentists(ctx)
	if err != nil || len(allDentists) == 0 {
		return "Sorry, I'm having trouble accessing the dentist list.", nil
	}

	dentistList := allDentists
	if a.Dentist != "" {
		dentistList = []string{a.Dentist}
	}

	var available []struct{ Dentist, Time string }

	for _, dentist := range dentistList {
		did, err := d.pg.GetDentistID(ctx, dentist)
		if err != nil {
			continue
		}
		booked, err := d.pg.GetBookedSlots(ctx, a.Date, &did)
		if err != nil {
			continue
		}
		bookedTimes := make(map[string]bool)
		for _, b := range booked {
			bookedTimes[b.Time] = true
		}
		for _, slot := range slots {
			if !bookedTimes[slot] {
				available = append(available, struct{ Dentist, Time string }{dentist, slot})
			}
		}
	}

	if len(available) == 0 {
		msg := fmt.Sprintf("No available slots on %s.", a.Date)
		if a.Dentist != "" {
			msg = fmt.Sprintf("No available slots on %s with %s.", a.Date, a.Dentist)
		}
		return msg + " Would you like to try another date?", nil
	}

	limit := 5
	if len(available) < limit {
		limit = len(available)
	}

	var parts []string
	for i := 0; i < limit; i++ {
		t, _ := time.Parse("15:04", available[i].Time)
		parts = append(parts, fmt.Sprintf("%s with %s", t.Format("3:04 PM"), available[i].Dentist))
	}

	result := fmt.Sprintf("Available slots on %s: %s.", a.Date, strings.Join(parts, ", "))
	if len(available) > 5 {
		result += fmt.Sprintf(" (%d more available)", len(available)-5)
	}
	return result, nil
}

// ─── book_appointment ───────────────────────────────────────────

func (d *ToolDispatcher) bookAppointment(ctx context.Context, args json.RawMessage) (string, error) {
	var a struct {
		PatientName  string `json:"patientName"`
		PatientPhone string `json:"patientPhone"`
		PatientEmail string `json:"patientEmail"`
		Date         string `json:"date"`
		Time         string `json:"time"`
		Dentist      string `json:"dentist"`
		Service      string `json:"service"`
		Notes        string `json:"notes"`
	}
	json.Unmarshal(args, &a)

	// Validate required fields
	if a.PatientName == "" || a.Date == "" || a.Time == "" || a.Dentist == "" || a.Service == "" {
		return "I need the patient name, date, time, dentist, and service to book. Please provide all required details.", nil
	}

	// Check availability again before booking
	did, err := d.pg.GetDentistID(ctx, a.Dentist)
	if err != nil {
		return fmt.Sprintf("I couldn't find dentist: %s", a.Dentist), nil
	}

	taken, err := d.pg.IsSlotTaken(ctx, a.Date, a.Time, did)
	if err != nil {
		return "I'm having trouble checking availability. Please try again.", nil
	}
	if taken {
		return fmt.Sprintf("The slot %s at %s with %s is already booked. Please choose another time.", a.Date, a.Time, a.Dentist), nil
	}

	// Find or create patient
	pid, err := d.pg.FindOrCreatePatient(ctx, a.PatientName, a.PatientPhone, strPtr(a.PatientEmail), nil)
	if err != nil {
		return fmt.Sprintf("Error registering patient: %v", err), nil
	}

	// Book the appointment
	notes := strPtr(a.Notes)
	if a.Notes == "" {
		notes = nil
	}
	aptID, err := d.pg.CreateAppointment(ctx, pid, a.Date, a.Time, a.Service, did, notes)
	if err != nil {
		// Check for unique constraint violation (double-booking race condition)
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return fmt.Sprintf("That slot was just booked by someone else. Please choose another time.", a.Date, a.Time), nil
		}
		return fmt.Sprintf("Error booking appointment: %v", err), nil
	}

	t, _ := time.Parse("15:04", a.Time)
	return fmt.Sprintf("Appointment booked successfully! Confirmation ID: %s. %s on %s at %s with %s for %s.",
		aptID, a.PatientName, a.Date, t.Format("3:04 PM"), a.Dentist, a.Service), nil
}

// ─── cancel_appointment ─────────────────────────────────────────

func (d *ToolDispatcher) cancelAppointment(ctx context.Context, args json.RawMessage) (string, error) {
	var a struct {
		PatientName string `json:"patientName"`
		Date        string `json:"date"`
	}
	json.Unmarshal(args, &a)
	if a.PatientName == "" {
		return "I need the patient name to cancel an appointment.", nil
	}
	if a.Date == "" {
		return "I need the appointment date (YYYY-MM-DD) to cancel.", nil
	}
	result, err := d.pg.CancelAppointment(ctx, a.Date, a.PatientName)
	if err != nil {
		return "Error cancelling appointment. Please call the clinic directly.", nil
	}
	return result, nil
}

// ─── parse_date ─────────────────────────────────────────────────

// ordinalWords maps English ordinal words to their numeric values
var ordinalWords = map[string]int{
	"first": 1, "second": 2, "third": 3, "fourth": 4, "fifth": 5,
	"sixth": 6, "seventh": 7, "eighth": 8, "ninth": 9, "tenth": 10,
	"eleventh": 11, "twelfth": 12, "thirteenth": 13, "fourteenth": 14,
	"fifteenth": 15, "sixteenth": 16, "seventeenth": 17, "eighteenth": 18,
	"nineteenth": 19, "twentieth": 20, "twenty-first": 21, "twenty-second": 22,
	"twenty-third": 23, "twenty-fourth": 24, "twenty-fifth": 25,
	"twenty-sixth": 26, "twenty-seventh": 27, "twenty-eighth": 28,
	"twenty-ninth": 29, "thirtieth": 30, "thirty-first": 31,
}

// dayOfWeekWords maps day names to time.Weekday
var dayOfWeekWords = map[string]time.Weekday{
	"sunday": time.Sunday, "monday": time.Monday, "tuesday": time.Tuesday,
	"wednesday": time.Wednesday, "thursday": time.Thursday, "friday": time.Friday,
	"saturday": time.Saturday,
}

func parseDate(args json.RawMessage) (string, error) {
	text := parseArg(args, "date_text")
	if text == "" {
		text = parseArg(args, "dateText")
	}
	if text == "" {
		return "Please provide a date to parse.", nil
	}

	now := time.Now()
	loc, _ := time.LoadLocation("America/Toronto")
	today := now.In(loc).Truncate(24 * time.Hour)
	lower := strings.ToLower(strings.TrimSpace(text))

	// ── Relative date words ──
	switch lower {
	case "today":
		return today.Format("2006-01-02"), nil
	case "tomorrow":
		return today.AddDate(0, 0, 1).Format("2006-01-02"), nil
	case "yesterday":
		return today.AddDate(0, 0, -1).Format("2006-01-02"), nil
	case "the day after tomorrow":
		return today.AddDate(0, 0, 2).Format("2006-01-02"), nil
	}

	// ── Chinese relative dates ──
	if strings.ContainsAny(lower, "\u4e00\u4e8c\u4e09\u56db\u4e94\u516d\u4e03\u516b\u4e5d\u5341\u767e\u5343\u4e07\u660e\u540e\u5927\u524d\u4eca") {
		switch {
		case lower == "今天" || lower == "今天":
			return today.Format("2006-01-02"), nil
		case lower == "明天" || lower == "明日":
			return today.AddDate(0, 0, 1).Format("2006-01-02"), nil
		case lower == "后天":
			return today.AddDate(0, 0, 2).Format("2006-01-02"), nil
		case strings.HasPrefix(lower, "下周一") || strings.HasPrefix(lower, "下星期一"):
			return nextWeekday(time.Monday, today).Format("2006-01-02"), nil
		case strings.HasPrefix(lower, "下周二") || strings.HasPrefix(lower, "下星期二"):
			return nextWeekday(time.Tuesday, today).Format("2006-01-02"), nil
		case strings.HasPrefix(lower, "下周三") || strings.HasPrefix(lower, "下星期三"):
			return nextWeekday(time.Wednesday, today).Format("2006-01-02"), nil
		case strings.HasPrefix(lower, "下周四") || strings.HasPrefix(lower, "下星期四"):
			return nextWeekday(time.Thursday, today).Format("2006-01-02"), nil
		case strings.HasPrefix(lower, "下周五") || strings.HasPrefix(lower, "下星期五"):
			return nextWeekday(time.Friday, today).Format("2006-01-02"), nil
		case strings.HasPrefix(lower, "下周六") || strings.HasPrefix(lower, "下星期六"):
			return nextWeekday(time.Saturday, today).Format("2006-01-02"), nil
		case strings.HasPrefix(lower, "这周一") || strings.HasPrefix(lower, "本周一"):
			return thisWeekday(time.Monday, today).Format("2006-01-02"), nil
		case strings.HasPrefix(lower, "这周二") || strings.HasPrefix(lower, "本周二"):
			return thisWeekday(time.Tuesday, today).Format("2006-01-02"), nil
		case strings.HasPrefix(lower, "这周三") || strings.HasPrefix(lower, "本周三"):
			return thisWeekday(time.Wednesday, today).Format("2006-01-02"), nil
		case strings.HasPrefix(lower, "这周四") || strings.HasPrefix(lower, "本周四"):
			return thisWeekday(time.Thursday, today).Format("2006-01-02"), nil
		case strings.HasPrefix(lower, "这周五") || strings.HasPrefix(lower, "本周五"):
			return thisWeekday(time.Friday, today).Format("2006-01-02"), nil
		}
	}

	// ── "next Monday", "this Friday", "next week Monday" ──
	for dayWord, weekday := range dayOfWeekWords {
		// "next Monday", "next monday"
		if strings.HasPrefix(lower, "next "+dayWord) {
			return nextWeekday(weekday, today).Format("2006-01-02"), nil
		}
		// "this Monday", "this monday"
		if strings.HasPrefix(lower, "this "+dayWord) {
			return thisWeekday(weekday, today).Format("2006-01-02"), nil
		}
	}

	// ── "next week", "next month" ──
	if lower == "next week" {
		return today.AddDate(0, 0, 7).Format("2006-01-02"), nil
	}
	if lower == "next month" {
		return today.AddDate(0, 1, 0).Format("2006-01-02"), nil
	}

	// ── Clean up the text: strip ordinal suffixes from NUMBERS only ──
	// "13th" → "13", "1st" → "1", but NOT "thirteenth"
	cleaned := lower
	for _, suffix := range []string{"th", "st", "nd", "rd"} {
		cleaned = cleanedSuffixReplacer(cleaned, suffix)
	}
	cleaned = strings.ReplaceAll(cleaned, ",", "")
	fields := strings.Fields(cleaned)

	// ── Try structured date formats ──
	// "April 13 2026", "april 13 2026", "april 13", "13 april 2026"
	monthMap := map[string]time.Month{
		"january": time.January, "jan": time.January,
		"february": time.February, "feb": time.February,
		"march": time.March, "mar": time.March,
		"april": time.April, "apr": time.April,
		"may": time.May,
		"june": time.June, "jun": time.June,
		"july": time.July, "jul": time.July,
		"august": time.August, "aug": time.August,
		"september": time.September, "sep": time.September,
		"october": time.October, "oct": time.October,
		"november": time.November, "nov": time.November,
		"december": time.December, "dec": time.December,
	}

	// Try standard Go formats first (YYYY-MM-DD, etc.)
	formats := []string{"2006-01-02", "January 2, 2006", "Jan 2, 2006", "2006/01/02", "01/02/2006", "January 2 2006", "Jan 2 2006", "2 Jan 2006", "2 January 2006"}
	for _, f := range formats {
		if t, err := time.ParseInLocation(f, text, loc); err == nil {
			return t.Format("2006-01-02"), nil
		}
		// Also try without comma
		if strings.Contains(f, ",") {
			noComma := strings.Replace(f, ",", "", 1)
			if t, err := time.ParseInLocation(noComma, text, loc); err == nil {
				return t.Format("2006-01-02"), nil
			}
		}
	}

	// Try "Month Day Year" from fields
	if len(fields) >= 2 {
		var month time.Month
		var day, year int
		var consumed int

		// Case: "april 13 2026" → month=april, day=13, year=2026
		if m, ok := monthMap[fields[0]]; ok {
			month = m
			day = atoi(fields[1])
			if day == 0 {
				// Try ordinal word: "april thirteenth"
				day = ordinalWords[fields[1]]
			}
			if len(fields) >= 3 {
				year = atoi(fields[2])
				consumed = 3
			} else {
				year = inferYear(month, day, today)
				consumed = 2
			}
		} else if m, ok := monthMap[strings.Join(fields[:2], " ")]; ok && len(fields) >= 3 {
			// "january 1 2026" but fields[0] could be "january"
			month = m
			day = atoi(fields[2])
			year = inferYear(month, day, today)
			consumed = 3
		} else if day = atoi(fields[0]); day > 0 && len(fields) >= 2 {
			// Case: "13 april 2026"
			if m, ok := monthMap[fields[1]]; ok {
				month = m
				if len(fields) >= 3 {
					year = atoi(fields[2])
				} else {
					year = inferYear(month, day, today)
				}
				consumed = 3
			}
		}

		if month != 0 && day > 0 && year > 0 && consumed > 0 {
			t := time.Date(year, month, day, 0, 0, 0, 0, loc)
			return t.Format("2006-01-02"), nil
		}
	}

	// Try "Month Day" with ordinal words: "april thirteenth"
	if len(fields) >= 2 {
		if m, ok := monthMap[fields[0]]; ok {
			day := 0
			if d := atoi(fields[1]); d > 0 {
				day = d
			} else if d, ok := ordinalWords[fields[1]]; ok {
				day = d
			}
			if day > 0 {
				year := inferYear(m, day, today)
				t := time.Date(year, m, day, 0, 0, 0, 0, loc)
				return t.Format("2006-01-02"), nil
			}
		}
	}

	return fmt.Sprintf("I couldn't parse '%s'. Please use YYYY-MM-DD format.", text), nil
}

// nextWeekday returns the next occurrence of the given weekday (must be in the future)
func nextWeekday(weekday time.Weekday, today time.Time) time.Time {
	daysAhead := int(weekday) - int(today.Weekday())
	if daysAhead <= 0 {
		daysAhead += 7
	}
	return today.AddDate(0, 0, daysAhead)
}

// thisWeekday returns the given weekday in the current week (may be today or past)
func thisWeekday(weekday time.Weekday, today time.Time) time.Time {
	daysAhead := int(weekday) - int(today.Weekday())
	if daysAhead < 0 {
		daysAhead += 7
	}
	if daysAhead == 0 && today.Weekday() != weekday {
		daysAhead = 7
	}
	return today.AddDate(0, 0, daysAhead)
}

// inferYear guesses the year based on context
func inferYear(month time.Month, day int, today time.Time) int {
	year := today.Year()
	candidate := time.Date(year, month, day, 0, 0, 0, 0, today.Location())
	// If the candidate is in the past and more than 30 days ago, assume next year
	if candidate.Before(today.AddDate(0, 0, -30)) {
		year++
	}
	return year
}

func atoi(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}

// cleanedSuffixReplacer strips ordinal suffixes only from number tokens
// e.g. "13th" → "13", but "thirteenth" stays intact
func cleanedSuffixReplacer(s, suffix string) string {
	parts := strings.Fields(s)
	for i, part := range parts {
		if len(part) > 2 && strings.HasSuffix(part, suffix) {
			prefix := strings.TrimSuffix(part, suffix)
			// Only strip if the prefix is purely numeric
			isNum := true
			for _, c := range prefix {
				if c < '0' || c > '9' {
					isNum = false
					break
				}
			}
			if isNum {
				parts[i] = prefix
			}
		}
	}
	return strings.Join(parts, " ")
}

// ─── get_next_available_dates ────────────────────────────────────

func (d *ToolDispatcher) getNextAvailableDates(ctx context.Context, args json.RawMessage) (string, error) {
	var a struct {
		Days string `json:"days"`
	}
	json.Unmarshal(args, &a)
	lookAhead := 14
	if a.Days != "" {
		fmt.Sscanf(a.Days, "%d", &lookAhead)
	}

	allDentists, err := d.pg.GetAllDentists(ctx)
	if err != nil {
		return "I'm having trouble checking availability.", nil
	}

	today := time.Now()
	loc, _ := time.LoadLocation("America/Toronto")
	today = today.In(loc).Truncate(24 * time.Hour)

	var openDates []string
	for i := 1; i <= lookAhead && len(openDates) < 5; i++ {
		date := today.AddDate(0, 0, i)
		dow := date.Weekday()
		if dow == time.Sunday {
			continue
		}
		dateStr := date.Format("2006-01-02")

		did, err := d.pg.GetDentistID(ctx, allDentists[0])
		if err != nil {
			continue
		}
		booked, err := d.pg.GetBookedSlots(ctx, dateStr, &did)
		if err != nil {
			continue
		}
		bookedCount := len(booked)
		slots := weekdaySlots
		if dow == time.Saturday {
			slots = saturdaySlots
		}
		if bookedCount < len(slots) {
			openDates = append(openDates, dateStr)
		}
	}

	if len(openDates) == 0 {
		return "No available dates found in the next two weeks.", nil
	}
	return fmt.Sprintf("Next available dates: %s.", strings.Join(openDates, ", ")), nil
}

// ─── Stub implementations (no DB needed) ────────────────────────

func sendBookingConfirmation(args json.RawMessage) (string, error) {
	return `{"confirmation": "Your appointment has been confirmed. You will receive a text message with the details."}`, nil
}

func validatePatientInfo(args json.RawMessage) (string, error) {
	name := parseArg(args, "patient_name")
	phone := parseArg(args, "patient_phone")
	email := parseArg(args, "patient_email")
	noMobile := parseArg(args, "no_mobile")

	if name == "" {
		return `{"valid": false, "missing": ["patient_name"], "message": "I need your full name to book the appointment."}`, nil
	}
	if phone == "" && noMobile != "true" && email == "" {
		return `{"valid": false, "missing": ["patient_phone"], "message": "I need your cell phone number for confirmation."}`, nil
	}
	if noMobile == "true" && email == "" {
		return `{"valid": false, "missing": ["patient_email"], "message": "Since you don't have a mobile, I need your email for confirmation."}`, nil
	}
	return `{"valid": true, "message": "All required information collected."}`, nil
}

func isBookingComplete(args json.RawMessage) (string, error) {
	var ctx struct {
		Service     string `json:"service"`
		Dentist     string `json:"dentist"`
		Date        string `json:"date"`
		Time        string `json:"time"`
		PatientName string `json:"patient_name"`
		PatientPhone string `json:"patient_phone"`
		PatientEmail string `json:"patient_email"`
	}
	json.Unmarshal(args, &ctx)

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

	if len(missing) > 0 {
		return fmt.Sprintf(`{"complete": false, "missing": ["%s"]}`, strings.Join(missing, `", "`)), nil
	}
	return `{"complete": true, "missing": []}`, nil
}

func getBookingStep(args json.RawMessage) (string, error) {
	var ctx struct {
		Service string `json:"service"`
		Dentist string `json:"dentist"`
		Date    string `json:"date"`
		Time    string `json:"time"`
		Name    string `json:"patient_name"`
		Phone   string `json:"patient_phone"`
	}
	json.Unmarshal(args, &ctx)

	if ctx.Service == "" {
		return `{"step": "collect_service", "message": "What dental service do you need today?"}`, nil
	}
	if ctx.Dentist == "" {
		return `{"step": "collect_dentist", "message": "Do you have a preferred dentist?"}`, nil
	}
	if ctx.Date == "" {
		return `{"step": "collect_date", "message": "What date would you like the appointment?"}`, nil
	}
	if ctx.Time == "" {
		return `{"step": "collect_time", "message": "What time works best for you?"}`, nil
	}
	if ctx.Name == "" {
		return `{"step": "collect_name", "message": "May I have your full name?"}`, nil
	}
	if ctx.Phone == "" {
		return `{"step": "collect_phone", "message": "What is your cell phone number for confirmation?"}`, nil
	}
	return `{"step": "confirm_booking", "message": "Please confirm your appointment details."}`, nil
}

func fillBookingFields(args json.RawMessage) (string, error) {
	// Echo back the args as the filled context
	return string(args), nil
}

func getConfirmMessage(args json.RawMessage) (string, error) {
	var ctx struct {
		Name    string `json:"patient_name"`
		Service string `json:"service"`
		Dentist string `json:"dentist"`
		Date    string `json:"date"`
		Time    string `json:"time"`
	}
	json.Unmarshal(args, &ctx)
	t, _ := time.Parse("15:04", ctx.Time)
	return fmt.Sprintf(`{"message": "Please confirm: %s, %s with %s on %s at %s. Is this correct?"}`,
		ctx.Name, ctx.Service, ctx.Dentist, ctx.Date, t.Format("3:04 PM")), nil
}

func getCancelMessage(args json.RawMessage) (string, error) {
	confirmed := parseArg(args, "confirmed")
	if confirmed == "true" {
		return `{"message": "Your appointment has been cancelled. Is there anything else I can help with?"}`, nil
	}
	return `{"message": "No problem. Your appointment remains scheduled."}`, nil
}

func getRescheduleMessage(args json.RawMessage) (string, error) {
	hasExisting := parseArg(args, "has_existing")
	if hasExisting == "true" {
		return `{"message": "I see you have an existing appointment. Let me help you reschedule it."}`, nil
	}
	return `{"message": "I'd be happy to help you reschedule. What new date and time works for you?"}`, nil
}

func getEmergencyMessage(args json.RawMessage) (string, error) {
	return `{"message": "For dental emergencies, please call 911 or visit the nearest emergency room immediately. For severe pain, knocked-out teeth, or heavy bleeding, do not wait."}`, nil
}

func detectLanguage(args json.RawMessage) (string, error) {
	sentence := parseArg(args, "sentence")
	if sentence == "" {
		return `{"error": "sentence is required"}`, nil
	}
	// Simple heuristic: if it contains Chinese characters, return zh
	for _, r := range sentence {
		if r >= '\u4e00' && r <= '\u9fff' {
			return `{"lang_code": "zh", "first_response": "您好！请问有什么可以帮您的？", "confidence": 0.95}`, nil
		}
	}
	return `{"lang_code": "en", "first_response": "Hello! How can I help you today?", "confidence": 0.95}`, nil
}

func classifyIntent(args json.RawMessage) (string, error) {
	utterance := parseArg(args, "utterance")
	if utterance == "" {
		return `{"error": "utterance is required"}`, nil
	}
	lower := strings.ToLower(utterance)
	intent := "general"
	if strings.Contains(lower, "book") || strings.Contains(lower, "appointment") {
		intent = "booking"
	} else if strings.Contains(lower, "cancel") {
		intent = "cancellation"
	} else if strings.Contains(lower, "hour") || strings.Contains(lower, "open") || strings.Contains(lower, "close") {
		intent = "hours"
	} else if strings.Contains(lower, "dentist") || strings.Contains(lower, "doctor") {
		intent = "dentists"
	}
	return fmt.Sprintf(`{"intent": "%s", "confidence": 0.8, "entities": []}`, intent), nil
}

// ─── Helpers ─────────────────────────────────────────────────────

func parseArg(args json.RawMessage, key string) string {
	var m map[string]interface{}
	json.Unmarshal(args, &m)
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
