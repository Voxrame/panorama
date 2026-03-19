use glam::{Vec2, Vec3};

pub mod obj;

#[derive(Debug, Default)]
pub struct Model {
    pub meshes: Vec<Mesh>,
}

#[derive(Debug, Default)]
pub struct Mesh {
    pub vertices: Vec<Vertex>,
}

#[derive(Debug, Clone, Copy)]
pub struct Vertex {
    pub position: Vec3,
    pub normal: Vec3,
    pub texcoord: Vec2,
}

impl Default for Vertex {
    fn default() -> Self {
        Self {
            position: Vec3::ZERO,
            normal: Vec3::ZERO,
            texcoord: Vec2::ZERO,
        }
    }
}
