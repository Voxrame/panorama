use std::net::SocketAddr;

use anyhow::{Context, Result};
use deadpool_postgres::{Manager, ManagerConfig, Pool, RecyclingMethod};
use panorama_config::Config;
use tokio_postgres::Config as PostgresConfig;
use tokio_postgres::NoTls;
use tracing::info;

use panorama_web::controller::Server;

#[tokio::main]
async fn main() -> Result<()> {
    tracing_subscriber::fmt().init();

    let config = Config::open("config.yaml").context("Can't open config")?;

    let addr: SocketAddr = config
        .web
        .listen_address
        .parse()
        .context("Invalid listen address")?;

    let db_pool = create_postgres_pool(&config.db.dsn)
        .context("Failed to create database connection pool")?;

    let server = Server::new(addr, db_pool);

    info!("Ready. http://{}", config.web.listen_address);

    server.run().await?;

    Ok(())
}

fn create_postgres_pool(dsn: &str) -> Result<Pool> {
    let pg_config = dsn
        .parse::<PostgresConfig>()
        .with_context(|| format!("Failed to parse DSN: {dsn}"))?;

    let mgr_config = ManagerConfig {
        recycling_method: RecyclingMethod::Fast,
    };

    let mgr = Manager::from_config(pg_config, NoTls, mgr_config);
    let pool = Pool::builder(mgr)
        .max_size(16)
        .build()
        .context("Failed to build pool")?;

    Ok(pool)
}
