use sqlx::PgPool;

use crate::config::AppConfig;

#[derive(Clone)]
pub struct PostgresDB {
    pool: PgPool,
}

impl PostgresDB {
    pub async fn connect() -> anyhow::Result<Self> {
        let config = AppConfig::get();
        let pool = PgPool::connect(&config.database_url).await?;
        tracing::info!("✅ PostgreSQL connected");
        Ok(Self { pool })
    }
}
