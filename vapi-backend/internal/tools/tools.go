package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
	_ "time/tzdata" // embed timezone DB so Alpine containers can load locations

	"dental-ai-vapi/internal/db"
	"dental-ai-vapi/internal/util"
)

var weekdaySlots = []string{"08:00", "08:30", "09:00", "09:30", "10:00", "10:30", "11:00", "11:30", "12:00", "12:30", "13:00", "13:30", "14:00", "14:30", "15:00", "15:30", "16:00", "16:30", "17:00", "17:30"}
var saturdaySlots = []string{"09:00", "09:30", "10:00", "10:30", "11:00", "11:30", "12:30", "13:00", "13:30"}

func CheckAvailability(ctx context.Context, pg *db.Postgres, args json.RawMessage) (string, string) {
	date := util.ParseArg(args, "date")
	dentist := util.ParseArg(args, "dentist")

	requestedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return fmt.Sprintf("Invalid date format: %s. Please use YYYY-MM-DD.", date), "error"
	}

	today := time.Now().Truncate(24 * time.Hour)
	if requestedDate.Before(today) {
		return "That date is in the past. Could you provide a future date?", "error"
	}

	dow := requestedDate.Weekday()
	if dow == time.Sunday {
		return "Sorry, the clinic is closed on Sundays. We're open Monday–Friday 8 AM–6 PM and Saturday 9 AM–2 PM.", "error"
	}

	slots := weekdaySlots
	if dow == time.Saturday {
		slots = saturdaySlots
	}

	// Load dentists from DB so the list stays in sync with the database
	allDentists, err := pg.GetAllDentists(ctx)
	if err != nil || len(allDentists) == 0 {
		log.Printf("GetAllDentists error: %v", err)
		return "Sorry, I'm having trouble accessing the dentist list right now. Please try again.", "error"
	}

	dentistList := allDentists
	if dentist != "" {
		dentistList = []string{dentist}
	}

	var available []struct{ Date, Dentist, Time string }

	for _, d := range dentistList {
		did, err := pg.GetDentistID(ctx, d)
		if err != nil {
			continue
		}

		booked, err := pg.GetBookedSlots(ctx, date, &did)
		if err != nil {
			log.Printf("GetBookedSlots error: %v", err)
			continue
		}

		bookedTimes := make(map[string]bool)
		for _, b := range booked {
			bookedTimes[b.Time] = true
		}

		for _, slot := range slots {
			if !bookedTimes[slot] {
				available = append(available, struct{ Date, Dentist, Time string }{date, d, slot})
			}
		}
	}

	if len(available) == 0 {
		msg := fmt.Sprintf("No available slots on %s.", date)
		if dentist != "" {
			msg = fmt.Sprintf("No available slots on %s with %s.", date, dentist)
		}
		msg += " Would you like to try another date?"
		return msg, "error"
	}

	limit := 5
	if len(available) < limit {
		limit = len(available)
	}

	var parts []string
	for i := 0; i < limit; i++ {
		parts = append(parts, fmt.Sprintf("%s with %s", util.TimeString(util.MustParseTime(available[i].Time)), available[i].Dentist))
	}

	result := fmt.Sprintf("Available slots on %s: %s.", date, util.JoinStr(parts, ", "))
	if len(available) > 5 {
		result += fmt.Sprintf(" (%d more available)", len(available)-5)
	}
	result += " Which time works for you?"

	return result, "success"
}

func BookAppointment(ctx context.Context, pg *db.Postgres, mongo *db.Mongo, args json.RawMessage) (string, string) {
	patientName := util.ParseArg(args, "patientName")
	patientPhone := util.ParseArg(args, "patientPhone")
	patientEmail := util.ParseArgOpt(args, "patientEmail")
	date := util.ParseArg(args, "date")
	timeStr := util.ParseArg(args, "time")
	dentist := util.ParseArg(args, "dentist")
	service := util.ParseArg(args, "service")
	notes := util.ParseArgOpt(args, "notes")

	if patientName == "" || date == "" || timeStr == "" || dentist == "" {
		return "I'm missing some details. Could you confirm the patient name, date, time, and dentist?", "error"
	}

	if service == "" {
		service = "General Checkup"
	}

	patientID, err := pg.FindOrCreatePatient(ctx, patientName, patientPhone, patientEmail)
	if err != nil {
		log.Printf("FindOrCreatePatient error: %v", err)
		return "Sorry, I had trouble saving your information. Please try again.", "error"
	}

	dentistID, err := pg.GetDentistID(ctx, dentist)
	if err != nil {
		return fmt.Sprintf("Dentist '%s' not found.", dentist), "error"
	}

	appt, err := pg.CreateAppointment(ctx, patientID, dentistID, service, date, timeStr, notes)
	if err != nil {
		log.Printf("CreateAppointment error: %v", err)
		return "Sorry, I had trouble booking that appointment. The slot may be taken. Please try a different time.", "error"
	}

	emailNote := ""
	if patientEmail != nil {
		emailNote = fmt.Sprintf(" A confirmation email will be sent to %s.", *patientEmail)
	} else {
		emailNote = fmt.Sprintf(" We'll confirm by phone at %s.", patientPhone)
	}

	msg := fmt.Sprintf("Booked! %s has a %s appointment with %s on %s at %s. Appointment #%s.%s Please arrive 15 minutes early and bring your insurance card. Anything else?",
		appt.PatientName, appt.Service, appt.Dentist, appt.Date,
		util.TimeString(util.MustParseTime(appt.Time)), appt.ID, emailNote)

	return msg, "success"
}

func CancelAppointment(ctx context.Context, pg *db.Postgres, mongo *db.Mongo, args json.RawMessage) (string, string) {
	patientName := util.ParseArg(args, "patientName")
	date := util.ParseArg(args, "date")

	if patientName == "" || date == "" {
		return "I need the patient name and date to find the appointment. Could you provide those?", "error"
	}

	appt, err := pg.FindAppointment(ctx, patientName, date)
	if err != nil {
		return fmt.Sprintf("I couldn't find an appointment for %s on %s. Could you double-check the name and date?", patientName, date), "error"
	}

	apptID := 0
	fmt.Sscanf(appt.ID, "%d", &apptID)

	cancelled, err := pg.CancelAppointment(ctx, apptID)
	if err != nil || cancelled == nil {
		return "That appointment was already cancelled or not found.", "error"
	}

	msg := fmt.Sprintf("Cancelled! %s's %s on %s at %s has been removed. Would you like to rebook?",
		cancelled.PatientName, cancelled.Service, cancelled.Date,
		util.TimeString(util.MustParseTime(cancelled.Time)))

	return msg, "success"
}

func GetClinicInfo(ctx context.Context, mongo *db.Mongo, args json.RawMessage) (string, string) {
	topic := util.ParseArg(args, "topic")
	if topic == "" {
		topic = "general"
	}
	info := mongo.GetClinicInfo(ctx, topic)
	return info, "success"
}

func SendVerificationLink(ctx context.Context, mongo *db.Mongo, args json.RawMessage, baseURL string) (string, string) {
	email := util.ParseArg(args, "email")
	if email == "" || len(email) < 5 {
		return "That doesn't look like a valid email address. Could you spell it out again?", "error"
	}

	token, err := mongo.InitiateVerification(ctx, email)
	if err != nil {
		log.Printf("InitiateVerification error: %v", err)
		return "I wasn't able to send the verification email right now. Could you try a different email address?", "error"
	}

	_ = token // Token stored in MongoDB, verification link generated by /verify endpoint
	msg := fmt.Sprintf("I've sent a verification link to %s. Please check your inbox and click the link — it expires in 10 minutes. Once you've clicked it, let me know and I'll confirm the booking.", email)
	return msg, "success"
}

func CheckVerificationStatus(ctx context.Context, mongo *db.Mongo, args json.RawMessage) (string, string) {
	email := util.ParseArg(args, "email")
	if email == "" {
		return "No verification was initiated for that email. Please send a verification link first.", "error"
	}
	status := mongo.CheckVerification(ctx, email)
	switch status {
	case "verified":
		return "verified", "success"
	case "pending":
		return "The link hasn't been clicked yet. Please ask the patient to check their inbox and click the verification link, then ask me to check again.", "pending"
	default:
		return "The verification link has expired. Please send a new one using the send_verification_link tool.", "error"
	}
}

func GetDentists(ctx context.Context, pg *db.Postgres) (string, string) {
	names, err := pg.GetAllDentists(ctx)
	if err != nil || len(names) == 0 {
		return "I'm having trouble retrieving the dentist list right now.", "error"
	}
	return fmt.Sprintf("Our dentists are: %s.", util.JoinStr(names, ", ")), "success"
}

func GetCurrentDate() (string, string) {
	now := time.Now()
	loc, _ := time.LoadLocation("America/Toronto")
	now = now.In(loc)
	date := now.Format("Monday, January 2, 2006")
	timeStr := now.Format("3:04 PM")
	return fmt.Sprintf("Today is %s, %s Toronto time.", date, timeStr), "success"
}
