use serde::Serialize;

#[derive(Debug, Serialize)]
pub struct HealthResponse {
    pub status: String,
    pub service: String,
}

#[derive(Debug, Clone, Serialize)]
pub struct Appointment {
    pub id: String,
    pub patient_name: String,
    pub patient_phone: String,
    pub patient_email: Option<String>,
    pub date: String,
    pub time: String,
    pub dentist: String,
    pub service: String,
    pub notes: Option<String>,
    pub status: String,
    pub google_event_id: Option<String>,
    pub created_at: String,
    pub cancelled_at: Option<String>,
}
