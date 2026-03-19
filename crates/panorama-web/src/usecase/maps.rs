use anyhow::Result;
use deadpool_postgres::Pool;

use crate::domain::Map;
use crate::repo::MapsRepo;

pub struct MapsUsecase {
    maps_repo: MapsRepo,
}

impl MapsUsecase {
    pub fn new(db: Pool) -> Self {
        Self {
            maps_repo: MapsRepo::new(db),
        }
    }

    pub async fn list(&self, _request: ListRequest) -> Result<Vec<Map>> {
        self.maps_repo.list().await
    }
}

pub struct ListRequest {}
