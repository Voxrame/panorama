use anyhow::Result;

pub struct RgbaU8Image {
    data: Vec<u8>,
    width: u32,
    height: u32,
}

#[derive(Clone, Copy)]
pub struct RgbaU8 {
    pub r: u8,
    pub g: u8,
    pub b: u8,
    pub a: u8,
}

impl RgbaU8Image {
    pub fn new(width: u32, height: u32) -> Self {
        let size = (width * height * 4) as usize;
        Self {
            data: vec![0; size],
            width,
            height,
        }
    }

    pub fn get_pixel(&self, x: u32, y: u32) -> RgbaU8 {
        debug_assert!(x < self.width, "x out of bounds: {} >= {}", x, self.width);
        debug_assert!(y < self.height, "y out of bounds: {} >= {}", y, self.height);
        let idx = self.pixel_index(x, y);
        RgbaU8 {
            r: self.data[idx],
            g: self.data[idx + 1],
            b: self.data[idx + 2],
            a: self.data[idx + 3],
        }
    }

    pub fn set_pixel(&mut self, x: u32, y: u32, color: RgbaU8) {
        debug_assert!(x < self.width, "x out of bounds: {} >= {}", x, self.width);
        debug_assert!(y < self.height, "y out of bounds: {} >= {}", y, self.height);
        let idx = self.pixel_index(x, y);
        self.data[idx] = color.r;
        self.data[idx + 1] = color.g;
        self.data[idx + 2] = color.b;
        self.data[idx + 3] = color.a;
    }

    fn pixel_index(&self, x: u32, y: u32) -> usize {
        ((y * self.width + x) * 4) as usize
    }

    pub fn encode_to_png(&self) -> Result<Vec<u8>> {
        let mut output = Vec::new();
        {
            let mut encoder = png::Encoder::new(&mut output, self.width, self.height);
            encoder.set_color(png::ColorType::Rgba);
            encoder.set_depth(png::BitDepth::Eight);
            let mut writer = encoder.write_header()?;
            writer.write_image_data(&self.data)?;
        }
        Ok(output)
    }

    pub fn width(&self) -> u32 {
        self.width
    }

    pub fn height(&self) -> u32 {
        self.height
    }
}

//

#[derive(Clone, Copy)]
pub struct RgbaF32 {
    pub r: f32,
    pub g: f32,
    pub b: f32,
    pub a: f32,
}

pub struct RgbaF32Image {
    data: Vec<f32>,
    width: u32,
    height: u32,
}

impl RgbaF32Image {
    pub fn new(width: u32, height: u32) -> Self {
        let size = (width * height * 4) as usize;
        Self {
            data: vec![0.0; size],
            width,
            height,
        }
    }

    pub fn get_pixel(&self, x: u32, y: u32) -> RgbaF32 {
        debug_assert!(x < self.width, "x out of bounds: {} >= {}", x, self.width);
        debug_assert!(y < self.height, "y out of bounds: {} >= {}", y, self.height);
        let idx = self.pixel_index(x, y);
        RgbaF32 {
            r: self.data[idx],
            g: self.data[idx + 1],
            b: self.data[idx + 2],
            a: self.data[idx + 3],
        }
    }

    pub fn set_pixel(&mut self, x: u32, y: u32, color: RgbaF32) {
        debug_assert!(x < self.width, "x out of bounds: {} >= {}", x, self.width);
        debug_assert!(y < self.height, "y out of bounds: {} >= {}", y, self.height);
        let idx = self.pixel_index(x, y);
        self.data[idx] = color.r;
        self.data[idx + 1] = color.g;
        self.data[idx + 2] = color.b;
        self.data[idx + 3] = color.a;
    }

    fn pixel_index(&self, x: u32, y: u32) -> usize {
        ((y * self.width + x) * 4) as usize
    }

    pub fn convert_to_ldr(&self, tone_map: impl ToneMap) -> Result<RgbaU8Image> {
        let mut ldr_image = RgbaU8Image::new(self.width, self.height);

        for y in 0..self.height {
            for x in 0..self.width {
                let hdr_pixel = self.get_pixel(x, y);
                let ldr_pixel = tone_map.map_color(hdr_pixel);
                ldr_image.set_pixel(x, y, ldr_pixel);
            }
        }

        Ok(ldr_image)
    }
}

pub trait ToneMap {
    fn map_color(&self, input: RgbaF32) -> RgbaU8;
}

#[derive(Clone, Copy)]
pub struct LinearToneMap;

impl ToneMap for LinearToneMap {
    fn map_color(&self, input: RgbaF32) -> RgbaU8 {
        RgbaU8 {
            r: (input.r * 255.0).clamp(0.0, 255.0) as u8,
            g: (input.g * 255.0).clamp(0.0, 255.0) as u8,
            b: (input.b * 255.0).clamp(0.0, 255.0) as u8,
            a: (input.a * 255.0).clamp(0.0, 255.0) as u8,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    mod rgba_u8_image_tests {
        use super::*;

        #[test]
        fn test_new_creates_image_with_correct_dimensions() {
            let img = RgbaU8Image::new(10, 20);
            assert_eq!(img.width, 10);
            assert_eq!(img.height, 20);
            assert_eq!(img.data.len(), 10 * 20 * 4);
        }

        #[test]
        fn test_new_initializes_data_to_zero() {
            let img = RgbaU8Image::new(5, 5);
            assert!(img.data.iter().all(|&v| v == 0));
        }

        #[test]
        fn test_pixel_index_calculation() {
            let img = RgbaU8Image::new(10, 10);
            assert_eq!(img.pixel_index(0, 0), 0);
            assert_eq!(img.pixel_index(1, 0), 4);
            assert_eq!(img.pixel_index(0, 1), 40);
            assert_eq!(img.pixel_index(5, 3), 3 * 40 + 5 * 4);
        }

        #[test]
        fn test_set_and_get_pixel() {
            let mut img = RgbaU8Image::new(10, 10);
            let color = RgbaU8 {
                r: 255,
                g: 128,
                b: 64,
                a: 200,
            };

            img.set_pixel(5, 5, color);
            let retrieved = img.get_pixel(5, 5);

            assert_eq!(retrieved.r, 255);
            assert_eq!(retrieved.g, 128);
            assert_eq!(retrieved.b, 64);
            assert_eq!(retrieved.a, 200);
        }

        #[test]
        fn test_set_pixel_does_not_affect_other_pixels() {
            let mut img = RgbaU8Image::new(10, 10);
            let color = RgbaU8 {
                r: 255,
                g: 255,
                b: 255,
                a: 255,
            };

            img.set_pixel(5, 5, color);

            // Check surrounding pixels are still zero
            assert_eq!(img.get_pixel(4, 5).r, 0);
            assert_eq!(img.get_pixel(6, 5).r, 0);
            assert_eq!(img.get_pixel(5, 4).r, 0);
            assert_eq!(img.get_pixel(5, 6).r, 0);
        }

        #[test]
        fn test_encode_to_png_creates_valid_png() {
            let mut img = RgbaU8Image::new(2, 2);
            img.set_pixel(
                0,
                0,
                RgbaU8 {
                    r: 255,
                    g: 0,
                    b: 0,
                    a: 255,
                },
            );
            img.set_pixel(
                1,
                0,
                RgbaU8 {
                    r: 0,
                    g: 255,
                    b: 0,
                    a: 255,
                },
            );
            img.set_pixel(
                0,
                1,
                RgbaU8 {
                    r: 0,
                    g: 0,
                    b: 255,
                    a: 255,
                },
            );
            img.set_pixel(
                1,
                1,
                RgbaU8 {
                    r: 255,
                    g: 255,
                    b: 255,
                    a: 255,
                },
            );

            let png_data = img.encode_to_png().unwrap();

            // PNG signature: 137 80 78 71 13 10 26 10
            assert_eq!(&png_data[0..8], &[137, 80, 78, 71, 13, 10, 26, 10]);
            // IHDR chunk
            assert_eq!(&png_data[12..16], b"IHDR");
        }

        #[test]
        #[should_panic]
        fn test_get_pixel_out_of_bounds_x() {
            let img = RgbaU8Image::new(10, 10);
            // x=10 is out of bounds (valid range is 0-9)
            let _ = img.get_pixel(10, 5);
        }

        #[test]
        #[should_panic]
        fn test_get_pixel_out_of_bounds_y() {
            let img = RgbaU8Image::new(10, 10);
            // y=10 is out of bounds (valid range is 0-9)
            let _ = img.get_pixel(5, 10);
        }

        #[test]
        #[should_panic]
        fn test_set_pixel_out_of_bounds_x() {
            let mut img = RgbaU8Image::new(10, 10);
            let color = RgbaU8 {
                r: 255,
                g: 0,
                b: 0,
                a: 255,
            };
            img.set_pixel(10, 5, color);
        }

        #[test]
        #[should_panic]
        fn test_set_pixel_out_of_bounds_y() {
            let mut img = RgbaU8Image::new(10, 10);
            let color = RgbaU8 {
                r: 255,
                g: 0,
                b: 0,
                a: 255,
            };
            img.set_pixel(5, 10, color);
        }

        #[test]
        #[should_panic]
        fn test_get_pixel_zero_dimensions() {
            let img = RgbaU8Image::new(0, 0);
            let _ = img.get_pixel(0, 0);
        }
    }

    mod rgba_f32_image_tests {
        use super::*;

        #[test]
        fn test_new_creates_image_with_correct_dimensions() {
            let img = RgbaF32Image::new(10, 20);
            assert_eq!(img.width, 10);
            assert_eq!(img.height, 20);
            assert_eq!(img.data.len(), 10 * 20 * 4);
        }

        #[test]
        fn test_new_initializes_data_to_zero() {
            let img = RgbaF32Image::new(5, 5);
            assert!(img.data.iter().all(|&v| v == 0.0));
        }

        #[test]
        fn test_pixel_index_calculation() {
            let img = RgbaF32Image::new(10, 10);
            assert_eq!(img.pixel_index(0, 0), 0);
            assert_eq!(img.pixel_index(1, 0), 4);
            assert_eq!(img.pixel_index(0, 1), 40);
            assert_eq!(img.pixel_index(5, 3), 3 * 40 + 5 * 4);
        }

        #[test]
        fn test_set_and_get_pixel() {
            let mut img = RgbaF32Image::new(10, 10);
            let color = RgbaF32 {
                r: 1.0,
                g: 0.5,
                b: 0.25,
                a: 0.75,
            };

            img.set_pixel(5, 5, color);
            let retrieved = img.get_pixel(5, 5);

            assert!((retrieved.r - 1.0).abs() < f32::EPSILON);
            assert!((retrieved.g - 0.5).abs() < f32::EPSILON);
            assert!((retrieved.b - 0.25).abs() < f32::EPSILON);
            assert!((retrieved.a - 0.75).abs() < f32::EPSILON);
        }

        #[test]
        fn test_set_pixel_does_not_affect_other_pixels() {
            let mut img = RgbaF32Image::new(10, 10);
            let color = RgbaF32 {
                r: 1.0,
                g: 1.0,
                b: 1.0,
                a: 1.0,
            };

            img.set_pixel(5, 5, color);

            assert_eq!(img.get_pixel(4, 5).r, 0.0);
            assert_eq!(img.get_pixel(6, 5).r, 0.0);
            assert_eq!(img.get_pixel(5, 4).r, 0.0);
            assert_eq!(img.get_pixel(5, 6).r, 0.0);
        }

        #[test]
        fn test_convert_to_ldr_basic() {
            let mut img = RgbaF32Image::new(2, 2);
            img.set_pixel(
                0,
                0,
                RgbaF32 {
                    r: 1.0,
                    g: 0.0,
                    b: 0.0,
                    a: 1.0,
                },
            );
            img.set_pixel(
                1,
                0,
                RgbaF32 {
                    r: 0.0,
                    g: 1.0,
                    b: 0.0,
                    a: 1.0,
                },
            );
            img.set_pixel(
                0,
                1,
                RgbaF32 {
                    r: 0.0,
                    g: 0.0,
                    b: 1.0,
                    a: 1.0,
                },
            );
            img.set_pixel(
                1,
                1,
                RgbaF32 {
                    r: 0.5,
                    g: 0.5,
                    b: 0.5,
                    a: 0.5,
                },
            );

            let ldr = img.convert_to_ldr(LinearToneMap).unwrap();

            assert_eq!(ldr.get_pixel(0, 0).r, 255);
            assert_eq!(ldr.get_pixel(0, 0).g, 0);
            assert_eq!(ldr.get_pixel(0, 0).b, 0);

            assert_eq!(ldr.get_pixel(1, 0).r, 0);
            assert_eq!(ldr.get_pixel(1, 0).g, 255);
            assert_eq!(ldr.get_pixel(1, 0).b, 0);

            assert_eq!(ldr.get_pixel(0, 1).r, 0);
            assert_eq!(ldr.get_pixel(0, 1).g, 0);
            assert_eq!(ldr.get_pixel(0, 1).b, 255);

            assert_eq!(ldr.get_pixel(1, 1).r, 127);
            assert_eq!(ldr.get_pixel(1, 1).g, 127);
            assert_eq!(ldr.get_pixel(1, 1).b, 127);
            assert_eq!(ldr.get_pixel(1, 1).a, 127);
        }

        #[test]
        fn test_convert_to_ldr_clamps_values() {
            let mut img = RgbaF32Image::new(1, 1);
            img.set_pixel(
                0,
                0,
                RgbaF32 {
                    r: 2.0,
                    g: -0.5,
                    b: 1.5,
                    a: 1.0,
                },
            );

            let ldr = img.convert_to_ldr(LinearToneMap).unwrap();

            assert_eq!(ldr.get_pixel(0, 0).r, 255); // clamped to 255
            assert_eq!(ldr.get_pixel(0, 0).g, 0); // clamped to 0
            assert_eq!(ldr.get_pixel(0, 0).b, 255); // clamped to 255
        }

        #[test]
        fn test_convert_to_ldr_preserves_dimensions() {
            let img = RgbaF32Image::new(100, 50);
            let ldr = img.convert_to_ldr(LinearToneMap).unwrap();

            assert_eq!(ldr.width, 100);
            assert_eq!(ldr.height, 50);
        }

        #[test]
        #[should_panic]
        fn test_get_pixel_out_of_bounds_x() {
            let img = RgbaF32Image::new(10, 10);
            // x=10 is out of bounds (valid range is 0-9)
            let _ = img.get_pixel(10, 5);
        }

        #[test]
        #[should_panic]
        fn test_get_pixel_out_of_bounds_y() {
            let img = RgbaF32Image::new(10, 10);
            // y=10 is out of bounds (valid range is 0-9)
            let _ = img.get_pixel(5, 10);
        }

        #[test]
        #[should_panic]
        fn test_set_pixel_out_of_bounds_x() {
            let mut img = RgbaF32Image::new(10, 10);
            let color = RgbaF32 {
                r: 1.0,
                g: 0.0,
                b: 0.0,
                a: 1.0,
            };
            img.set_pixel(10, 5, color);
        }

        #[test]
        #[should_panic]
        fn test_set_pixel_out_of_bounds_y() {
            let mut img = RgbaF32Image::new(10, 10);
            let color = RgbaF32 {
                r: 1.0,
                g: 0.0,
                b: 0.0,
                a: 1.0,
            };
            img.set_pixel(5, 10, color);
        }

        #[test]
        #[should_panic]
        fn test_get_pixel_zero_dimensions() {
            let img = RgbaF32Image::new(0, 0);
            let _ = img.get_pixel(0, 0);
        }
    }

    mod tone_map_tests {
        use super::*;

        #[test]
        fn test_linear_tone_map_white() {
            let tone_map = LinearToneMap;
            let input = RgbaF32 {
                r: 1.0,
                g: 1.0,
                b: 1.0,
                a: 1.0,
            };
            let output = tone_map.map_color(input);

            assert_eq!(output.r, 255);
            assert_eq!(output.g, 255);
            assert_eq!(output.b, 255);
            assert_eq!(output.a, 255);
        }

        #[test]
        fn test_linear_tone_map_black() {
            let tone_map = LinearToneMap;
            let input = RgbaF32 {
                r: 0.0,
                g: 0.0,
                b: 0.0,
                a: 0.0,
            };
            let output = tone_map.map_color(input);

            assert_eq!(output.r, 0);
            assert_eq!(output.g, 0);
            assert_eq!(output.b, 0);
            assert_eq!(output.a, 0);
        }

        #[test]
        fn test_linear_tone_map_gray() {
            let tone_map = LinearToneMap;
            let input = RgbaF32 {
                r: 0.5,
                g: 0.5,
                b: 0.5,
                a: 0.5,
            };
            let output = tone_map.map_color(input);

            assert_eq!(output.r, 127);
            assert_eq!(output.g, 127);
            assert_eq!(output.b, 127);
            assert_eq!(output.a, 127);
        }

        #[test]
        fn test_linear_tone_map_clamps_above_one() {
            let tone_map = LinearToneMap;
            let input = RgbaF32 {
                r: 1.5,
                g: 2.0,
                b: -0.5,
                a: 1.0,
            };
            let output = tone_map.map_color(input);

            assert_eq!(output.r, 255);
            assert_eq!(output.g, 255);
            assert_eq!(output.b, 0);
            assert_eq!(output.a, 255);
        }

        #[test]
        fn test_linear_tone_map_partial_values() {
            let tone_map = LinearToneMap;
            let input = RgbaF32 {
                r: 0.25,
                g: 0.5,
                b: 0.75,
                a: 1.0,
            };
            let output = tone_map.map_color(input);

            assert_eq!(output.r, 63);
            assert_eq!(output.g, 127);
            assert_eq!(output.b, 191);
            assert_eq!(output.a, 255);
        }
    }
}
