use anyhow::{Context, Result};
use glam::IVec3;
use rusqlite::{Connection, OptionalExtension, params};

use crate::Backend;

pub struct SQLite3Backend {
    conn: Connection,
}

impl SQLite3Backend {
    pub fn new(path: &str) -> Result<Self> {
        let conn = Connection::open(path)
            .with_context(|| format!("failed to open sqlite database at {}", path))?;

        Ok(Self { conn })
    }
}

impl Backend for SQLite3Backend {
    fn get_block_data(&self, position: IVec3) -> Result<Option<Vec<u8>>> {
        let mut stmt = self
            .conn
            .prepare("SELECT data FROM blocks WHERE x=$1 AND y=$2 AND z=$3")?;

        let data: Option<Vec<u8>> = stmt
            .query_row(params![position.x, position.y, position.z], |row| {
                row.get(0)
            })
            .optional()?;

        Ok(data)
    }
}
