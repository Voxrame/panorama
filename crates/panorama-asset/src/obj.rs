use anyhow::{Context, Result, bail};
use std::io::BufRead;

use super::{Mesh, Model, Vertex};

fn parse_vector3(fields: &[&str]) -> Result<[f32; 3]> {
    if fields.len() < 3 {
        bail!(
            "expected at least 3 vector elements, found {}",
            fields.len()
        );
    }

    let x = fields[0].parse::<f32>().context("parse x")?;
    let y = fields[1].parse::<f32>().context("parse y")?;
    let z = fields[2].parse::<f32>().context("parse z")?;

    Ok([x, y, z])
}

fn parse_vector2(fields: &[&str]) -> Result<[f32; 2]> {
    if fields.len() < 2 {
        bail!(
            "expected at least 2 vector elements, found {}",
            fields.len()
        );
    }

    let x = fields[0].parse::<f32>().context("parse x")?;
    let y = fields[1].parse::<f32>().context("parse y")?;

    Ok([x, y])
}

#[derive(Debug, Clone, Copy)]
struct Triplet {
    position_index: usize,
    texcoord_index: Option<usize>,
    normal_index: Option<usize>,
}

fn parse_face(fields: &[&str]) -> Result<Vec<Triplet>> {
    if fields.len() < 3 {
        bail!("face needs at least 3 vertices, got {}", fields.len());
    }

    let mut triplets = Vec::new();

    for field in fields {
        let parts: Vec<&str> = field.split('/').collect();

        let position_index = parts[0].parse::<usize>().context("parse position index")?;

        let texcoord_index = if parts.len() > 1 && !parts[1].is_empty() {
            Some(parts[1].parse::<usize>().context("parse texcoord index")?)
        } else {
            None
        };

        let normal_index = if parts.len() > 2 && !parts[2].is_empty() {
            Some(parts[2].parse::<usize>().context("parse normal index")?)
        } else {
            None
        };

        triplets.push(Triplet {
            position_index,
            texcoord_index,
            normal_index,
        });
    }

    Ok(triplets)
}

#[derive(Default)]
struct ObjParser {
    positions: Vec<[f32; 3]>,
    texcoords: Vec<[f32; 2]>,
    normals: Vec<[f32; 3]>,
    meshes: Vec<Mesh>,
    current_mesh: Mesh,
}

impl ObjParser {
    fn vertex_at(&self, triplet: Triplet) -> Vertex {
        let texcoord = triplet
            .texcoord_index
            .map(|i| self.texcoords[i - 1])
            .unwrap_or([0.0, 0.0]);
        let normal = triplet
            .normal_index
            .map(|i| self.normals[i - 1])
            .unwrap_or([0.0, 0.0, 0.0]);

        Vertex {
            position: self.positions[triplet.position_index - 1].into(),
            normal: normal.into(),
            texcoord: texcoord.into(),
        }
    }

    fn triangulate_polygon(&self, triplets: &[Triplet]) -> Vec<Vertex> {
        let mut vertices = Vec::new();

        let origin = self.vertex_at(triplets[0]);

        for i in 2..triplets.len() {
            vertices.push(origin);
            vertices.push(self.vertex_at(triplets[i - 1]));
            vertices.push(self.vertex_at(triplets[i]));
        }

        vertices
    }

    fn start_new_mesh(&mut self) {
        if !self.current_mesh.vertices.is_empty() {
            self.meshes.push(std::mem::take(&mut self.current_mesh));
        }
    }

    fn process_line(&mut self, line: &str) -> Result<()> {
        let trimmed = line.trim();

        if trimmed.starts_with('#') || trimmed.is_empty() {
            return Ok(());
        }

        let fields: Vec<&str> = trimmed.split_whitespace().collect();
        if fields.is_empty() {
            return Ok(());
        }

        match fields[0] {
            "v" => {
                let position = parse_vector3(&fields[1..]).context("parse vertex position")?;
                self.positions.push(position);
            }
            "vt" => {
                let texcoord = parse_vector2(&fields[1..]).context("parse texture coordinate")?;
                self.texcoords.push(texcoord);
            }
            "vn" => {
                let normal = parse_vector3(&fields[1..]).context("parse vertex normal")?;
                self.normals.push(normal);
            }
            "f" => {
                let triplets = parse_face(&fields[1..]).context("parse face")?;
                let vertices = self.triangulate_polygon(&triplets);
                self.current_mesh.vertices.extend(vertices);
            }
            "g" => {
                self.start_new_mesh();
            }
            "o" => {
                self.start_new_mesh();
            }
            _ => {}
        }

        Ok(())
    }

    fn finish(mut self) -> Model {
        if !self.current_mesh.vertices.is_empty() {
            self.meshes.push(self.current_mesh);
        }
        Model {
            meshes: self.meshes,
        }
    }
}

pub fn load_obj(data: &[u8]) -> Result<Model> {
    let mut parser = ObjParser::default();

    for (line_num, line) in data.lines().enumerate() {
        let line = line.context("read line")?;
        parser
            .process_line(&line)
            .with_context(|| format!("parse line {}", line_num + 1))?;
    }

    Ok(parser.finish())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_simple_triangle() {
        let obj_content = r#"
v 0.0 0.0 0.0
v 1.0 0.0 0.0
v 0.5 1.0 0.0
f 1 2 3
"#;
        let model = load_obj(obj_content.as_bytes()).unwrap();

        assert_eq!(model.meshes.len(), 1);
        assert_eq!(model.meshes[0].vertices.len(), 3);

        assert_eq!(model.meshes[0].vertices[0].position, [0.0, 0.0, 0.0].into());
        assert_eq!(model.meshes[0].vertices[1].position, [1.0, 0.0, 0.0].into());
        assert_eq!(model.meshes[0].vertices[2].position, [0.5, 1.0, 0.0].into());
    }

    #[test]
    fn test_parse_with_texcoords_and_normals() {
        let obj_content = r#"
v 0.0 0.0 0.0
v 1.0 0.0 0.0
v 0.5 1.0 0.0
vt 0.0 0.0
vt 1.0 0.0
vt 0.5 1.0
vn 0.0 0.0 1.0
f 1/1/1 2/2/1 3/3/1
"#;
        let model = load_obj(obj_content.as_bytes()).unwrap();

        assert_eq!(model.meshes.len(), 1);
        assert_eq!(model.meshes[0].vertices.len(), 3);

        assert_eq!(model.meshes[0].vertices[0].texcoord, [0.0, 0.0].into());
        assert_eq!(model.meshes[0].vertices[1].texcoord, [1.0, 0.0].into());
        assert_eq!(model.meshes[0].vertices[2].texcoord, [0.5, 1.0].into());

        assert_eq!(model.meshes[0].vertices[0].normal, [0.0, 0.0, 1.0].into());
    }

    #[test]
    fn test_parse_quad_triangulation() {
        let obj_content = r#"
v 0.0 0.0 0.0
v 1.0 0.0 0.0
v 1.0 1.0 0.0
v 0.0 1.0 0.0
f 1 2 3 4
"#;
        let model = load_obj(obj_content.as_bytes()).unwrap();

        assert_eq!(model.meshes.len(), 1);
        assert_eq!(model.meshes[0].vertices.len(), 6);
    }

    #[test]
    fn test_parse_multiple_groups() {
        let obj_content = r#"
v 0.0 0.0 0.0
v 1.0 0.0 0.0
v 0.5 1.0 0.0

g group1
f 1 2 3

g group2
f 1 2 3
"#;
        let model = load_obj(obj_content.as_bytes()).unwrap();

        assert_eq!(model.meshes.len(), 2);
        assert_eq!(model.meshes[0].vertices.len(), 3);
        assert_eq!(model.meshes[1].vertices.len(), 3);
    }

    #[test]
    fn test_parse_multiple_objects() {
        let obj_content = r#"
v 0.0 0.0 0.0
v 1.0 0.0 0.0
v 0.5 1.0 0.0

o object1
f 1 2 3

o object2
f 1 2 3
"#;
        let model = load_obj(obj_content.as_bytes()).unwrap();

        assert_eq!(model.meshes.len(), 2);
        assert_eq!(model.meshes[0].vertices.len(), 3);
        assert_eq!(model.meshes[1].vertices.len(), 3);
    }

    #[test]
    fn test_parse_with_comments() {
        let obj_content = r#"
# This is a comment
v 0.0 0.0 0.0
# Another comment
v 1.0 0.0 0.0
v 0.5 1.0 0.0
f 1 2 3
"#;
        let model = load_obj(obj_content.as_bytes()).unwrap();

        assert_eq!(model.meshes.len(), 1);
        assert_eq!(model.meshes[0].vertices.len(), 3);
    }

    #[test]
    fn test_parse_empty_lines() {
        let obj_content = r#"
v 0.0 0.0 0.0

v 1.0 0.0 0.0

v 0.5 1.0 0.0

f 1 2 3
"#;
        let model = load_obj(obj_content.as_bytes()).unwrap();

        assert_eq!(model.meshes.len(), 1);
        assert_eq!(model.meshes[0].vertices.len(), 3);
    }
}
