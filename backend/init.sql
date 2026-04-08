-- ─────────────────────────────────────────────────────────────
-- Dental AI Reception — PostgreSQL Schema
-- Runs once on first container creation
-- ─────────────────────────────────────────────────────────────

-- Patients table
CREATE TABLE IF NOT EXISTS patients (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    phone       VARCHAR(50) NOT NULL,
    email       VARCHAR(255),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_patients_phone ON patients(phone);
CREATE INDEX IF NOT EXISTS idx_patients_email ON patients(email);

-- Dentists table
CREATE TABLE IF NOT EXISTS dentists (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL UNIQUE,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed default dentists
INSERT INTO dentists (name) VALUES
    ('Dr. Sarah Chen'),
    ('Dr. Michael Park'),
    ('Dr. Priya Sharma')
ON CONFLICT (name) DO NOTHING;

-- Appointments table
CREATE TABLE IF NOT EXISTS appointments (
    id              SERIAL PRIMARY KEY,
    patient_id      INTEGER NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    dentist_id      INTEGER NOT NULL REFERENCES dentists(id) ON DELETE RESTRICT,
    service         VARCHAR(255) NOT NULL,
    appointment_date DATE NOT NULL,
    appointment_time TIME NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'confirmed',
    notes           TEXT,
    google_event_id VARCHAR(500),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    cancelled_at    TIMESTAMPTZ,

    CONSTRAINT chk_status CHECK (status IN ('confirmed', 'cancelled', 'no_show')),
    CONSTRAINT uq_dentist_datetime UNIQUE (dentist_id, appointment_date, appointment_time)
);

CREATE INDEX IF NOT EXISTS idx_appointments_date ON appointments(appointment_date);
CREATE INDEX IF NOT EXISTS idx_appointments_status ON appointments(status);
CREATE INDEX IF NOT EXISTS idx_appointments_patient ON appointments(patient_id);

-- Email verifications table
CREATE TABLE IF NOT EXISTS email_verifications (
    id              SERIAL PRIMARY KEY,
    email           VARCHAR(255) NOT NULL,
    token           VARCHAR(128) NOT NULL UNIQUE,
    verified        BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    verified_at     TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '10 minutes')
);

CREATE INDEX IF NOT EXISTS idx_email_verifications_email ON email_verifications(email);
CREATE INDEX IF NOT EXISTS idx_email_verifications_token ON email_verifications(token);

-- Clinic settings (configurable key-value store)
CREATE TABLE IF NOT EXISTS clinic_settings (
    key     VARCHAR(100) PRIMARY KEY,
    value   TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed default settings
INSERT INTO clinic_settings (key, value) VALUES
    ('clinic_name', 'Dental Clinic'),
    ('clinic_phone', '(555) 555-0123'),
    ('clinic_address', '123 Main Street'),
    ('clinic_city', 'City'),
    ('clinic_province', 'Province'),
    ('clinic_postal_code', 'A1A 1A1'),
    ('clinic_hours', 'Monday through Friday 8 AM to 6 PM, Saturday 9 AM to 2 PM. Closed on Sundays and statutory holidays.'),
    ('clinic_insurance', 'We accept most major dental insurance plans. Please bring your insurance card to your appointment.'),
    ('clinic_emergency', 'For after-hours dental emergencies, please go to your nearest emergency department.'),
    ('clinic_cancellation_policy', 'We ask for at least 24 hours notice for cancellations. Late cancellations or no-shows may be subject to a fee.'),
    ('clinic_new_patient_info', 'New patients are always welcome! For your first visit, please arrive 15 minutes early to complete paperwork. Bring your insurance card, a list of current medications, and any recent dental X-rays if you have them.'),
    ('clinic_payment', 'We accept cash, debit, Visa, Mastercard, and e-transfer. We also offer payment plans for larger treatments.')
ON CONFLICT (key) DO NOTHING;

-- Auto-update updated_at trigger for patients
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_patients_updated_at
    BEFORE UPDATE ON patients
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
