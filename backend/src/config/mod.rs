use std::sync::OnceLock;

pub static CONFIG: OnceLock<AppConfig> = OnceLock::new();

#[derive(Clone, Debug)]
pub struct AppConfig {
    pub port: u16,
    pub database_url: String,
    pub mongo_url: String,
    pub mongo_db: String,
    pub smtp_email: String,
    pub smtp_password: String,
}

impl AppConfig {
    pub fn from_env() -> Self {
        dotenvy::dotenv().ok();

        Self {
            port: std::env::var("PORT")
                .ok()
                .and_then(|p| p.parse().ok())
                .unwrap_or(3001),
            database_url: std::env::var("DATABASE_URL")
                .unwrap_or_else(|_| "postgres://dental:internal_pg_2024@postgres:5432/dental".to_string()),
            mongo_url: std::env::var("MONGO_URL")
                .unwrap_or_else(|_| "mongodb://mongo:27017/dental".to_string()),
            mongo_db: std::env::var("MONGO_DB")
                .unwrap_or_else(|_| "dental".to_string()),
            smtp_email: std::env::var("SMTP_EMAIL").unwrap_or_default(),
            smtp_password: std::env::var("SMTP_PASSWORD").unwrap_or_default(),
        }
    }

    pub fn get() -> &'static Self {
        CONFIG.get_or_init(|| AppConfig::from_env())
    }
}
