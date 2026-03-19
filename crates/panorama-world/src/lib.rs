pub mod block;
pub mod sqlite3;

use anyhow::{Context, Result, anyhow};
use glam::IVec3;
use std::collections::HashMap;

use crate::block::MapBlock;

pub struct World {
    backend: Box<dyn Backend>,
}

impl World {
    pub fn new(backend: impl Backend) -> Self {
        Self {
            backend: Box::new(backend),
        }
    }

    pub fn get_block(&self, position: IVec3) -> Result<Option<MapBlock>> {
        self.backend
            .get_block_data(position)
            .context("get mapblock data")?
            .map(|block| MapBlock::decode(&block).context("decode mapblock"))
            .transpose()
    }
}

pub trait Backend: 'static {
    fn get_block_data(&self, position: IVec3) -> Result<Option<Vec<u8>>>;
}

pub struct WorldMeta {
    data: HashMap<String, String>,
}

impl WorldMeta {
    pub fn parse(input: &[u8]) -> Result<Self> {
        let mut data = HashMap::new();

        let text = str::from_utf8(input).context("parse text as utf-8")?;

        for line in text.lines() {
            let line = line.trim();
            if line.is_empty() || line.starts_with("#") {
                continue;
            }

            let Some((key, value)) = line.split_once("=") else {
                return Err(anyhow!("invalid line: {}", line));
            };

            data.insert(key.trim().to_string(), value.trim().to_string());
        }

        Ok(Self { data })
    }

    pub fn get(&self, key: &str) -> Option<&str> {
        self.data.get(key).map(|value| value.as_str())
    }
}
