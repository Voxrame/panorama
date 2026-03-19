pub mod maps;

use std::net::SocketAddr;
use std::sync::Arc;

use axum::{routing::get, Router};
use deadpool_postgres::Pool;

use crate::usecase::maps::{MapsUsecase};

#[derive(Clone)]
pub struct AppState {
    maps_usecase: Arc<MapsUsecase>,
}

pub struct Server {
    addr: SocketAddr,
    db: Pool,
}

impl Server {
    pub fn new(addr: SocketAddr, db: Pool) -> Self {
        Self { addr, db }
    }

    pub async fn run(self) -> anyhow::Result<()> {
        let state = AppState {
            maps_usecase: Arc::new(MapsUsecase::new(self.db)),
        };

        let app = Router::new()
            .route("/api/v1/maps", get(maps::list))
            .with_state(state);

        let m = "a";

        let listener = tokio::net::TcpListener::bind(self.addr).await?;
        axum::serve(listener, app).await?;

        Ok(())
    }
}
