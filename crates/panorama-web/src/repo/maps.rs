use anyhow::Result;
use deadpool_postgres::Pool;

use crate::domain::Map;

pub struct MapsRepo {
    db: Pool,
}

impl MapsRepo {
    pub fn new(db: Pool) -> Self {
        Self { db }
    }

    pub async fn list(&self) -> Result<Vec<Map>> {
        let sql = r#"
            SELECT id, name, kind
            FROM map
        "#;

        let client = self.db.get().await?;
        let rows = client.query(sql, &[]).await?;

        let maps = rows
            .iter()
            .map(|row| Map {
                id: row.get("id"),
                name: row.get("name"),
                kind: row.get("kind"),
            })
            .collect();

        Ok(maps)
    }
}
