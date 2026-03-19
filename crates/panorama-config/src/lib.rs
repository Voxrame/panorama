use std::path::Path;

use anyhow::{Context, Result};
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize)]
pub struct Config {
    pub web: Web,
    pub db: Database,
}

#[derive(Serialize, Deserialize)]
pub struct Web {
    pub listen_address: String,
}

#[derive(Serialize, Deserialize)]
pub struct Database {
    pub dsn: String,
}

impl Config {
    pub fn open(path: impl AsRef<Path>) -> Result<Self> {
        let path = path.as_ref();

        let data =
            std::fs::read_to_string(path).with_context(|| format!("Failed to read {path:?}"))?;
        let config = serde_saphyr::from_str(&data).context("Failed to deserialize yaml")?;

        Ok(config)
    }
}
