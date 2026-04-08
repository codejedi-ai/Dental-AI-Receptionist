use mongodb::{bson, Client};
use mongodb::bson::doc;

use crate::config::AppConfig;

#[derive(Clone)]
pub struct MongoDB {
    client: Client,
    db_name: String,
}

fn bson_dt(dt: chrono::DateTime<chrono::Utc>) -> bson::DateTime {
    bson::DateTime::from_millis(dt.timestamp_millis())
}

impl MongoDB {
    pub async fn connect() -> anyhow::Result<Self> {
        let config = AppConfig::get();
        let client = Client::with_uri_str(&config.mongo_url).await?;
        client.database("admin").run_command(doc! { "ping": 1 }).await?;
        tracing::info!("✅ MongoDB connected");
        Ok(Self {
            client,
            db_name: config.mongo_db.clone(),
        })
    }
}
