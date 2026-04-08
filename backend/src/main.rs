use axum::{routing::get, Json, Router};
use tower_http::cors::CorsLayer;

mod config;
mod db;
mod models;
mod services;

use config::AppConfig;
use models::HealthResponse;

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    tracing_subscriber::fmt()
        .with_env_filter(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "info".into()),
        )
        .init();

    let config = AppConfig::get();
    tracing::info!("🚀 Starting dental-ai-browser-backend v{}", env!("CARGO_PKG_VERSION"));
    tracing::info!("   Port: {}", config.port);

    // Connect to databases (shared with Go Vapi backend)
    let _pg = db::postgres_db::PostgresDB::connect().await?;
    let _mongo = db::mongo_db::MongoDB::connect().await?;

    let mailer = services::email::create_mailer();
    let _ = mailer; // Used for email sending endpoints

    let app = Router::new()
        .route("/health", get(|| async {
            Json(HealthResponse {
                status: "ok".to_string(),
                service: "dental-ai-browser-backend".to_string(),
            })
        }))
        .route("/api/health", get(|| async {
            Json(HealthResponse {
                status: "ok".to_string(),
                service: "dental-ai-browser-backend".to_string(),
            })
        }))
        .layer(CorsLayer::permissive());

    let addr = format!("0.0.0.0:{}", config.port);
    tracing::info!("🦷 Browser backend listening on {}", addr);

    let listener = tokio::net::TcpListener::bind(&addr).await?;
    axum::serve(listener, app).await?;

    Ok(())
}
