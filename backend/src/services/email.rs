use lettre::message::{header::ContentType, Attachment, Mailbox, Message, MultiPart, SinglePart};
use lettre::transport::smtp::authentication::Credentials;
use lettre::{AsyncSmtpTransport, AsyncTransport, Tokio1Executor};

use crate::config::AppConfig;
use crate::models::Appointment;

pub type Mailer = AsyncSmtpTransport<Tokio1Executor>;

pub fn create_mailer() -> Mailer {
    let config = AppConfig::get();
    let creds = Credentials::new(config.smtp_email.clone(), config.smtp_password.clone());

    AsyncSmtpTransport::<Tokio1Executor>::relay("smtp.gmail.com")
        .unwrap()
        .credentials(creds)
        .build()
}

pub fn build_ics_content(appt: &Appointment) -> String {
    let local_start = format!("{}T{}:00-05:00", appt.date, appt.time);
    let start_dt = chrono::DateTime::parse_from_rfc3339(&local_start)
        .unwrap_or_else(|_| chrono::Utc::now().into());
    let end_dt = start_dt + chrono::Duration::minutes(30);

    let fmt = |dt: chrono::DateTime<chrono::FixedOffset>| {
        dt.format("%Y%m%dT%H%M%SZ").to_string()
    };

    let uid = format!(
        "{}-{}-{}@dental.local",
        appt.date,
        appt.time,
        appt.patient_name.replace(' ', "")
    );

    let description = format!(
        "Patient: {}\\nDentist: {}\\nService: {}{}",
        appt.patient_name,
        appt.dentist,
        appt.service,
        appt.notes
            .as_ref()
            .map(|n| format!("\\nNotes: {}", n))
            .unwrap_or_default()
    );

    let clinic = AppConfig::get();
    format!(
        "BEGIN:VCALENDAR\r\n\
        VERSION:2.0\r\n\
        PRODID:-//Dental Clinic//EN\r\n\
        METHOD:REQUEST\r\n\
        BEGIN:VEVENT\r\n\
        UID:{uid}\r\n\
        DTSTAMP:{now}\r\n\
        DTSTART:{start}\r\n\
        DTEND:{end}\r\n\
        SUMMARY:{service} — Dental Clinic\r\n\
        DESCRIPTION:{description}\r\n\
        LOCATION:{location}\r\n\
        ORGANIZER;CN=\"Dental Clinic\":mailto:{organizer}\r\n\
        ATTENDEE;CN=\"{patient}\";RSVP=TRUE:mailto:{attendee}\r\n\
        STATUS:CONFIRMED\r\n\
        SEQUENCE:0\r\n\
        BEGIN:VALARM\r\n\
        TRIGGER:-PT24H\r\n\
        ACTION:EMAIL\r\n\
        DESCRIPTION:Reminder: Dental appointment tomorrow\r\n\
        END:VALARM\r\n\
        BEGIN:VALARM\r\n\
        TRIGGER:-PT1H\r\n\
        ACTION:DISPLAY\r\n\
        DESCRIPTION:Dental appointment in 1 hour\r\n\
        END:VALARM\r\n\
        END:VEVENT\r\n\
        END:VCALENDAR",
        uid = uid,
        now = fmt(chrono::Utc::now().into()),
        start = fmt(start_dt),
        end = fmt(end_dt),
        service = appt.service,
        description = description,
        location = format!("{}, {}", clinic.smtp_email.split('@').next().unwrap_or("Clinic"), "Address"),
        organizer = clinic.smtp_email,
        patient = appt.patient_name,
        attendee = appt.patient_email.as_deref().unwrap_or("")
    )
}

pub async fn send_confirmation_email(mailer: &Mailer, appt: &Appointment) -> anyhow::Result<()> {
    let config = AppConfig::get();
    let Some(email_addr) = &appt.patient_email else {
        return Ok(());
    };

    if config.smtp_email.is_empty() || config.smtp_password.is_empty() {
        tracing::warn!("⚠️  SMTP credentials not set — skipping confirmation email");
        return Ok(());
    }

    let ics_content = build_ics_content(appt);

    let body_text = format!(
        "Hi {},\n\n\
        Your appointment has been confirmed!\n\n\
          Date:    {}\n\
          Time:    {}\n\
          Dentist: {}\n\
          Service: {}{}\n\n\
        Please arrive 15 minutes early and bring your insurance card.\n\
        The calendar invite is attached — click it to add to your calendar.\n\n\
        — Dental Clinic",
        appt.patient_name,
        appt.date,
        appt.time,
        appt.dentist,
        appt.service,
        appt.notes
            .as_ref()
            .map(|n| format!("\n  Notes:   {}", n))
            .unwrap_or_default()
    );

    let ics_attachment = Attachment::new("appointment.ics".to_string())
        .body(ics_content.into_bytes(), ContentType::parse("text/calendar; method=REQUEST; charset=utf-8").unwrap());

    let msg = Message::builder()
        .from(
            format!("Dental Clinic <{}>", config.smtp_email)
                .parse()
                .unwrap(),
        )
        .to(Mailbox::new(
            Some(appt.patient_name.clone()),
            email_addr.parse().unwrap(),
        ))
        .subject("Appointment Confirmed — Dental Clinic")
        .multipart(
            MultiPart::mixed()
                .singlepart(SinglePart::plain(body_text))
                .singlepart(ics_attachment),
        )?;

    mailer.send(msg).await?;
    tracing::info!("📧 Confirmation + .ics sent to {}", email_addr);
    Ok(())
}

pub async fn send_verification_email(mailer: &Mailer, email_addr: &str, verify_url: &str) -> anyhow::Result<()> {
    let config = AppConfig::get();

    if config.smtp_email.is_empty() || config.smtp_password.is_empty() {
        tracing::warn!("⚠️  SMTP credentials not set — skipping verification email");
        return Ok(());
    }

    let body_text = format!(
        "Hi there,\n\n\
        Please click the link below to verify your email address so we can confirm your appointment:\n\n\
        {}\n\n\
        This link expires in 10 minutes.\n\n\
        If you didn't request this, you can ignore this email.\n\n\
        — Dental Clinic",
        verify_url
    );

    let body_html = format!(
        r#"<!DOCTYPE html>
<html>
<body style="font-family:sans-serif;max-width:480px;margin:40px auto;color:#333">
  <h2 style="color:#1a6fc4">Dental Clinic</h2>
  <p>Hi there,</p>
  <p>Please verify your email address to confirm your appointment booking:</p>
  <a href="{verify_url}" style="display:inline-block;padding:12px 24px;background:#1a6fc4;color:#fff;border-radius:6px;text-decoration:none;font-weight:bold;margin:16px 0">
    Verify My Email
  </a>
  <p style="color:#666;font-size:13px">This link expires in 10 minutes.<br>If you didn't request this, ignore this email.</p>
  <hr style="border:none;border-top:1px solid #eee;margin-top:32px">
  <p style="color:#999;font-size:12px">Dental Clinic</p>
</body>
</html>"#,
        verify_url = verify_url
    );

    let msg = Message::builder()
        .from(
            format!("Dental Clinic <{}>", config.smtp_email)
                .parse()
                .unwrap(),
        )
        .to(email_addr.parse().unwrap())
        .subject("Verify your email — Dental Clinic")
        .multipart(
            MultiPart::alternative()
                .singlepart(SinglePart::plain(body_text))
                .singlepart(SinglePart::html(body_html)),
        )?;

    mailer.send(msg).await?;
    tracing::info!("📧 Verification email sent to {}", email_addr);
    Ok(())
}

pub async fn send_cancellation_email(mailer: &Mailer, appt: &Appointment) -> anyhow::Result<()> {
    let config = AppConfig::get();
    let Some(email_addr) = &appt.patient_email else {
        return Ok(());
    };

    if config.smtp_email.is_empty() || config.smtp_password.is_empty() {
        return Ok(());
    }

    let body_text = format!(
        "Hi {},\n\n\
        Your {} appointment with {} on {} at {} has been cancelled.\n\n\
        To rebook, please contact us.\n\n\
        — Dental Clinic",
        appt.patient_name,
        appt.service,
        appt.dentist,
        appt.date,
        appt.time
    );

    let msg = Message::builder()
        .from(
            format!("Dental Clinic <{}>", config.smtp_email)
                .parse()
                .unwrap(),
        )
        .to(email_addr.parse().unwrap())
        .subject("Appointment Cancelled — Dental Clinic")
        .body(body_text)?;

    mailer.send(msg).await?;
    tracing::info!("📧 Cancellation email sent to {}", email_addr);
    Ok(())
}
