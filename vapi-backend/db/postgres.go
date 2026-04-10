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

// Pool returns the underlying pgxpool.Pool for direct access
func (p *Postgres) Pool() *pgxpool.Pool {
	return p.pool
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

type Patient struct {
	ID          int32
	UUID        string
	Name        string
	Phone       string
	Email       *string
	Address     *string
	DateOfBirth *string
}

func (p *Postgres) GetPatientByPhone(ctx context.Context, phone string) (*Patient, error) {
	pt := &Patient{}
	err := p.pool.QueryRow(ctx,
		"SELECT id, uuid::text, name, phone, email, address, date_of_birth::text FROM patients WHERE phone = $1", phone).
		Scan(&pt.ID, &pt.UUID, &pt.Name, &pt.Phone, &pt.Email, &pt.Address, &pt.DateOfBirth)
	if err != nil {
		return nil, err
	}
	return pt, nil
}

func (p *Postgres) FindOrCreatePatient(ctx context.Context, name, phone string, email, address *string) (int32, error) {
	var id int32
	err := p.pool.QueryRow(ctx,
		`INSERT INTO patients (name, phone, email, address) VALUES ($1, $2, $3, $4)
		 ON CONFLICT (phone) DO UPDATE SET name = $1, email = $3, address = COALESCE($4, patients.address)
		 RETURNING id`, name, phone, email, address).Scan(&id)
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

func (p *Postgres) CreateAppointment(ctx context.Context, patientID int32, date, timeStr, service string, dentistID int32, notes *string) (string, error) {
	var id string
	err := p.pool.QueryRow(ctx,
		`INSERT INTO appointments (patient_id, appointment_date, appointment_time, dentist_id, service, status, notes)
		 VALUES ($1, $2, $3, $4, $5, 'confirmed', $6)
		 RETURNING id::text`,
		patientID, date, timeStr, dentistID, service, notes).Scan(&id)
	return id, err
}

func (p *Postgres) CancelAppointment(ctx context.Context, date string, patientName string) (string, error) {
	var count int64
	err := p.pool.QueryRow(ctx,
		`UPDATE appointments SET status = 'cancelled', cancelled_at = NOW()
		 FROM patients
		 WHERE appointments.patient_id = patients.id
		 AND patients.name ILIKE $1
		 AND appointments.appointment_date = $2
		 AND appointments.status = 'confirmed'`,
		"%"+patientName+"%", date).Scan(&count)
	if err != nil {
		return "", err
	}
	if count == 0 {
		return "No confirmed appointment found for that name and date.", nil
	}
	return fmt.Sprintf("Successfully cancelled appointment for %s on %s.", patientName, date), nil
}
