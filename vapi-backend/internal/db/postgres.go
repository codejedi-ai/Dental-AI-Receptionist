package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(ctx context.Context, dsn string) (*Postgres, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}
	return &Postgres{pool: pool}, nil
}

func (p *Postgres) Close() {
	p.pool.Close()
}

func (p *Postgres) GetDentistID(ctx context.Context, name string) (int32, error) {
	var id int32
	err := p.pool.QueryRow(ctx,
		"SELECT id FROM dentists WHERE name = $1 AND is_active = true", name).
		Scan(&id)
	return id, err
}

func (p *Postgres) GetAllDentists(ctx context.Context) ([]string, error) {
	rows, err := p.pool.Query(ctx, "SELECT name FROM dentists WHERE is_active = true ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		names = append(names, n)
	}
	return names, nil
}

func (p *Postgres) FindOrCreatePatient(ctx context.Context, name, phone string, email *string) (int32, error) {
	var id int32
	err := p.pool.QueryRow(ctx,
		`INSERT INTO patients (name, phone, email) VALUES ($1, $2, $3)
		 ON CONFLICT (phone) DO UPDATE SET name = $1, email = $3
		 RETURNING id`, name, phone, email).Scan(&id)
	return id, err
}

type BookedSlot struct {
	DentistID int32
	Time      string
}

func (p *Postgres) GetBookedSlots(ctx context.Context, date string, dentistID *int32) ([]BookedSlot, error) {
	var rows pgx.Rows
	var err error

	if dentistID != nil {
		rows, err = p.pool.Query(ctx,
			"SELECT dentist_id, appointment_time::text FROM appointments WHERE appointment_date = $1 AND status = 'confirmed' AND dentist_id = $2",
			date, *dentistID)
	} else {
		rows, err = p.pool.Query(ctx,
			"SELECT dentist_id, appointment_time::text FROM appointments WHERE appointment_date = $1 AND status = 'confirmed'",
			date)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slots []BookedSlot
	for rows.Next() {
		var s BookedSlot
		if err := rows.Scan(&s.DentistID, &s.Time); err != nil {
			return nil, err
		}
		// Normalize HH:MM:SS → HH:MM to match slot format
		if len(s.Time) > 5 {
			s.Time = s.Time[:5]
		}
		slots = append(slots, s)
	}
	return slots, nil
}

func (p *Postgres) IsSlotTaken(ctx context.Context, date, timeStr string, dentistID int32) (bool, error) {
	var count int64
	err := p.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM appointments WHERE appointment_date = $1 AND appointment_time = $2 AND dentist_id = $3 AND status = 'confirmed'",
		date, timeStr, dentistID).Scan(&count)
	return count > 0, err
}

type Appointment struct {
	ID            string  `json:"id"`
	PatientName   string  `json:"patientName"`
	PatientPhone  string  `json:"patientPhone"`
	PatientEmail  *string `json:"patientEmail,omitempty"`
	Date          string  `json:"date"`
	Time          string  `json:"time"`
	Dentist       string  `json:"dentist"`
	Service       string  `json:"service"`
	Notes         *string `json:"notes,omitempty"`
	Status        string  `json:"status"`
	GoogleEventID *string `json:"googleEventId,omitempty"`
	CreatedAt     string  `json:"createdAt"`
	CancelledAt   *string `json:"cancelledAt,omitempty"`
}

func (p *Postgres) CreateAppointment(ctx context.Context, patientID, dentistID int32, service, date, timeStr string, notes *string) (*Appointment, error) {
	var id int32
	err := p.pool.QueryRow(ctx,
		`INSERT INTO appointments (patient_id, dentist_id, service, appointment_date, appointment_time, notes)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id`, patientID, dentistID, service, date, timeStr, notes).Scan(&id)
	if err != nil {
		return nil, err
	}
	return p.getAppointment(ctx, int(id))
}

func (p *Postgres) FindAppointment(ctx context.Context, patientName, date string) (*Appointment, error) {
	var id int
	err := p.pool.QueryRow(ctx,
		`SELECT a.id FROM appointments a
		 JOIN patients p ON p.id = a.patient_id
		 WHERE a.appointment_date = $1 AND a.status = 'confirmed' AND LOWER(p.name) LIKE LOWER($2)
		 ORDER BY a.appointment_time LIMIT 1`,
		date, "%"+patientName+"%").Scan(&id)
	if err != nil {
		return nil, err
	}
	return p.getAppointment(ctx, id)
}

func (p *Postgres) CancelAppointment(ctx context.Context, id int) (*Appointment, error) {
	var affected int64
	err := p.pool.QueryRow(ctx,
		"UPDATE appointments SET status = 'cancelled', cancelled_at = NOW() WHERE id = $1 AND status = 'confirmed' RETURNING 1", id).
		Scan(&affected)
	if err != nil || affected == 0 {
		return nil, err
	}
	return p.getAppointment(ctx, id)
}

func (p *Postgres) getAppointment(ctx context.Context, id int) (*Appointment, error) {
	appt := &Appointment{ID: fmt.Sprintf("%d", id)}
	var patientID, dentistID int32
	err := p.pool.QueryRow(ctx,
		`SELECT patient_id, dentist_id, service, appointment_date::text, appointment_time::text,
			status, notes, created_at::text, cancelled_at::text, google_event_id
		 FROM appointments WHERE id = $1`, id).
		Scan(&patientID, &dentistID, &appt.Service, &appt.Date, &appt.Time,
			&appt.Status, &appt.Notes, &appt.CreatedAt, &appt.CancelledAt, &appt.GoogleEventID)
	if err != nil {
		return nil, err
	}

	err = p.pool.QueryRow(ctx, "SELECT name, phone, email FROM patients WHERE id = $1", patientID).
		Scan(&appt.PatientName, &appt.PatientPhone, &appt.PatientEmail)
	if err != nil {
		return nil, err
	}

	err = p.pool.QueryRow(ctx, "SELECT name FROM dentists WHERE id = $1", dentistID).
		Scan(&appt.Dentist)
	return appt, err
}
