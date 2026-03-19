use std::collections::HashMap;
use std::io::{Cursor, Read, Seek};

use anyhow::{Context, Result};
use flate2::bufread::ZlibDecoder;
use zstd::stream::read::Decoder as ZstdDecoder;

const BLOCK_SIZE: usize = 16;
const BLOCK_VOLUME: usize = BLOCK_SIZE * BLOCK_SIZE * BLOCK_SIZE;
const NODE_SIZE_IN_BYTES: usize = 4;

#[derive(Debug, Clone, Copy)]
pub struct Node {
    pub id: u16,
    pub param1: u8,
    pub param2: u8,
}

#[derive(Debug, Clone)]
pub struct MapBlock {
    mappings: HashMap<u16, String>,
    node_data: Vec<u8>,
}

impl MapBlock {
    pub fn decode(data: &[u8]) -> Result<Self> {
        let mut reader = Cursor::new(data);

        let version = read_u8(&mut reader)?;

        if version < 29 {
            let mapblock =
                Self::decode_legacy(&mut reader, version).context("decode legacy block")?;
            return Ok(mapblock);
        }

        Self::decode_v29_and_later(&mut reader)
    }

    pub fn resolve_name(&self, id: u16) -> Option<&str> {
        self.mappings.get(&id).map(|s| s.as_str())
    }

    pub fn get_node(&self, x: usize, y: usize, z: usize) -> Option<Node> {
        if x >= BLOCK_SIZE || y >= BLOCK_SIZE || z >= BLOCK_SIZE {
            return None;
        }

        let index = z * BLOCK_SIZE * BLOCK_SIZE + y * BLOCK_SIZE + x;

        let id_hi = self.node_data[2 * index] as u16;
        let id_lo = self.node_data[2 * index + 1] as u16;

        let param1 = self.node_data[2 * BLOCK_VOLUME + index];
        let param2 = self.node_data[3 * BLOCK_VOLUME + index];

        Some(Node {
            id: (id_hi << 8) | id_lo,
            param1,
            param2,
        })
    }
}

/// Serialization
impl MapBlock {
    fn decode_legacy(reader: &mut Cursor<&[u8]>, version: u8) -> Result<MapBlock> {
        if version >= 27 {
            // - uint8 flags
            // - uint16 lighting_complete
            // - uint8 content_width
            // - uint8 params_width
            reader
                .seek_relative(1 + 2 + 1 + 1)
                .context("seek past header (v27+)")?;
        } else {
            // - uint8 flags
            // - uint8 content_width
            // - uint8 params_width
            reader
                .seek_relative(1 + 1 + 1)
                .context("seek past header (pre-v27)")?;
        }

        // Read the current position to know where compressed data starts
        let pos = reader.position() as usize;
        let remaining = &reader.get_ref()[pos..];
        let node_data = inflate(remaining).context("inflate node data")?;

        // Skip the first inflate's input (we don't know exact size consumed)
        // Need to estimate: Zlib doesn't tell us exact consumed bytes easily
        // We'll try to read past it by creating a new decoder and checking
        let mut decoder = ZlibDecoder::new(remaining);
        let mut temp_buf = Vec::new();
        decoder.read_to_end(&mut temp_buf)?;
        let consumed = decoder.total_in() as usize;
        drop(decoder);

        // Seek past the consumed zlib data
        reader.seek_relative(consumed as i64)?;

        // Skip second inflate
        let pos = reader.position() as usize;
        let remaining = &reader.get_ref()[pos..];
        let mut decoder = ZlibDecoder::new(remaining);
        let mut temp_buf = Vec::new();
        decoder.read_to_end(&mut temp_buf)?;
        let consumed = decoder.total_in() as usize;
        drop(decoder);

        reader.seek_relative(consumed as i64)?;

        // - uint8 staticObjectVersion
        reader.seek_relative(1)?;

        let static_object_count = read_u16(reader)?;

        for _ in 0..static_object_count {
            // - uint8 type
            // - int32 x, y, z
            reader.seek_relative(1 + 4 + 4 + 4)?;

            let data_size = read_u16(reader)?;
            reader.seek_relative(data_size as i64)?;
        }

        // - uint32 timestamp
        // - uint8 mappingVersion
        reader.seek_relative(4 + 1)?;

        let mappings = read_mappings(reader)?;

        Ok(MapBlock {
            mappings,
            node_data,
        })
    }

    fn decode_v29_and_later(reader: &mut Cursor<&[u8]>) -> Result<MapBlock> {
        // The entire remaining data is zstd compressed
        let pos = reader.position() as usize;
        let compressed_data = &reader.get_ref()[pos..];

        let mut decoder = ZstdDecoder::new(compressed_data)?;
        let mut data = Vec::new();
        decoder.read_to_end(&mut data)?;

        let mut reader = Cursor::new(data.as_slice());

        // Skip:
        // - uint8 flags
        // - uint16 lighting_complete
        // - uint32 timestamp
        // - uint8 mapping version
        reader.seek_relative(1 + 2 + 4 + 1)?;

        let mappings = read_mappings(&mut reader)?;

        // Skip uint8 contentWidth, uint8 paramsWidth
        reader.seek_relative(1 + 1)?;

        let mut node_data = vec![0u8; BLOCK_VOLUME * NODE_SIZE_IN_BYTES];
        reader.read_exact(&mut node_data)?;

        Ok(MapBlock {
            mappings,
            node_data,
        })
    }
}

fn read_u8(reader: &mut Cursor<&[u8]>) -> Result<u8> {
    let mut buf = [0u8; 1];
    reader.read_exact(&mut buf)?;
    Ok(buf[0])
}

fn read_u16(reader: &mut Cursor<&[u8]>) -> Result<u16> {
    let mut buf = [0u8; 2];
    reader.read_exact(&mut buf)?;
    Ok(u16::from_be_bytes(buf))
}

fn read_string(reader: &mut Cursor<&[u8]>) -> Result<String> {
    let length = read_u16(reader)? as usize;
    let mut buf = vec![0u8; length];
    reader.read_exact(&mut buf)?;
    String::from_utf8(buf).context("invalid utf-8 string")
}

fn inflate(data: &[u8]) -> Result<Vec<u8>> {
    let mut decoder = ZlibDecoder::new(data);
    let mut result = Vec::new();
    decoder.read_to_end(&mut result)?;
    Ok(result)
}

fn read_mappings(reader: &mut Cursor<&[u8]>) -> Result<HashMap<u16, String>> {
    let mapping_count = read_u16(reader)?;
    let mut mappings = HashMap::new();

    for _ in 0..mapping_count {
        let id = read_u16(reader)?;
        let name = read_string(reader)?;
        mappings.insert(id, name);
    }

    Ok(mappings)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_block_size() {
        assert_eq!(BLOCK_VOLUME, 4096);
    }
}
