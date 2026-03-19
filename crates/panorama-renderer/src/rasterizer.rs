use glam::{Vec2, Vec3, Vec4, vec2, vec3, vec4};

use crate::image::RgbaU8Image;

pub struct Rasterizer {}

impl Rasterizer {
    pub fn new() -> Self {
        Self {}
    }
}

fn cartesian_to_barycentric(p: Vec2, a: Vec2, b: Vec2, c: Vec2) -> Vec3 {
    let u = vec3(c.x - a.x, b.x - a.x, a.x - p.x);
    let v = vec3(c.y - a.y, b.y - a.y, a.y - p.y);
    let w = u.cross(v);

    vec3(1.0 - (w.x + w.y) / w.z, w.y / w.z, w.x / w.z)
}

fn sample_triangle(x: i32, y: i32, a: Vec2, b: Vec2, c: Vec2) -> Vec3 {
    let p = vec2(x as f32, y as f32);
    let sample_point_offset = vec2(0.5, 0.5);
    let barycentric = cartesian_to_barycentric(p + sample_point_offset, a, b, c);

    barycentric
}

fn sample_texture(texture: &RgbaU8Image, texcoord: Vec2) -> Vec4 {
    let width = texture.width() as f32;
    let height = texture.height() as f32;

    let x = texcoord.x * width as f32;
    let y = texcoord.y * height as f32;

    if x < 0.0 || y < 0.0 || x + 1.0 > width || y + 1.0 > height {
        return vec4(1.0, 0.0, 1.0, 1.0);
    }

    let texel = texture.get_pixel(x as u32, y as u32);

    let r = texel.r as f32 / 255.0;
    let g = texel.g as f32 / 255.0;
    let b = texel.b as f32 / 255.0;
    let a = texel.a as f32 / 255.0;

    vec4(r, g, b, a)
}
