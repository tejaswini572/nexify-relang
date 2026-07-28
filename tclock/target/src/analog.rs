use std::collections::HashMap;
use std::f64::consts::PI;

#[derive(Clone, Copy, PartialEq, Eq)]
pub enum Color {
    Background,
    Hours,
    Minutes,
    Seconds,
}

pub struct VirtualGrid {
    pub pixels: HashMap<(i32, i32), Color>,
}

impl VirtualGrid {
    pub fn new() -> Self {
        Self {
            pixels: HashMap::new(),
        }
    }

    pub fn set_pixel(&mut self, x: i32, y: i32, color: Color) {
        self.pixels.insert((x, y), color);
    }

    pub fn draw_line(&mut self, sx: i32, sy: i32, x0: i32, y0: i32, color: Color) {
        let x1 = x0 + sx;
        let y0_virt = y0 * 2;
        let y1_virt = y0_virt + sy;

        let steep = (y1_virt - y0_virt).abs() > (x1 - x0).abs();
        let (mut start_x, mut start_y) = if steep { (y0_virt, x0) } else { (x0, y0_virt) };
        let (mut end_x, mut end_y) = if steep { (y1_virt, x1) } else { (x1, y1_virt) };

        if start_x > end_x {
            std::mem::swap(&mut start_x, &mut end_x);
            std::mem::swap(&mut start_y, &mut end_y);
        }

        let dx = end_x - start_x;
        let dy = (end_y - start_y).abs();
        let mut err = dx as f64 / 2.0;
        let y_step = if start_y < end_y { 1 } else { -1 };

        let mut y = start_y;
        for x in start_x..=end_x {
            if steep {
                self.set_pixel(y, x, color);
            } else {
                self.set_pixel(x, y, color);
            }
            err -= dy as f64;
            if err < 0.0 {
                y += y_step;
                err += dx as f64;
            }
        }
    }
}

fn angle_coords(max_v: f64, time_value: f64, radius: f64) -> (i32, i32) {
    let theta = 2.0 * PI * (max_v - time_value) / max_v;
    let x = (-theta.sin() * radius).round() as i32;
    let y = (-theta.cos() * radius).round() as i32;
    (x, y)
}

use chrono::{DateTime, Local, Timelike};

pub fn render_analog(time: DateTime<Local>, show_seconds: bool, radius: i32) -> Vec<String> {
    let r = radius as f64;
    let sec = time.second() as f64;
    let minute = time.minute() as f64;
    let hour = time.hour() as f64;

    let (sx, sy) = angle_coords(60.0, sec, 0.9 * r);
    let m = minute + sec / 60.0;
    let (mx, my) = angle_coords(60.0, m, 0.80 * r);
    let h = (hour % 12.0) + m / 60.0;
    let (hx, hy) = angle_coords(12.0, h, 0.47 * r);

    let mut grid = VirtualGrid::new();
    let cx = radius;
    let cy = radius / 2;

    if show_seconds {
        grid.draw_line(sx, sy, cx, cy, Color::Seconds);
    }
    grid.draw_line(mx, my, cx, cy, Color::Minutes);
    grid.draw_line(hx, hy, cx, cy, Color::Hours);

    // Setup rim chars
    let mut rim_chars: HashMap<(i32, i32), String> = HashMap::new();
    for n in 1..=60 {
        let (nx, ny) = angle_coords(60.0, (n % 60) as f64, r);
        let rx = cx + nx;
        let ry_phys = cy + ny / 2;
        let s = if n % 5 == 0 {
            (if n == 60 { 12 } else { n / 5 }).to_string()
        } else {
            "•".to_string()
        };
        for (i, c) in s.chars().enumerate() {
            rim_chars.insert((rx + i as i32, ry_phys), c.to_string());
        }
    }
    rim_chars.insert((cx, cy), "•".to_string()); // Center dot

    let width = radius * 2 + 1;
    let height = radius + 1;
    let mut lines = Vec::new();

    for y_phys in 0..height {
        let mut current_line = String::new();
        let y_virt_top = y_phys * 2;
        let y_virt_bottom = y_phys * 2 + 1;

        for x in 0..width {
            let top_c = grid.pixels.get(&(x, y_virt_top)).copied().unwrap_or(Color::Background);
            let bot_c = grid.pixels.get(&(x, y_virt_bottom)).copied().unwrap_or(Color::Background);

            if top_c == Color::Background && bot_c == Color::Background {
                if let Some(c) = rim_chars.get(&(x, y_phys)) {
                    // Dark gray text for rim marks
                    current_line.push_str(&format!("\x1b[38;2;140;140;140m{}\x1b[0m", c));
                } else {
                    current_line.push(' ');
                }
            } else if top_c == bot_c {
                current_line.push_str(&colorize(top_c, Color::Background, "█"));
            } else if top_c != Color::Background && bot_c == Color::Background {
                current_line.push_str(&colorize(top_c, Color::Background, "▀"));
            } else if top_c == Color::Background && bot_c != Color::Background {
                current_line.push_str(&colorize(bot_c, Color::Background, "▄"));
            } else {
                current_line.push_str(&colorize(bot_c, top_c, "▄"));
            }
        }
        lines.push(current_line);
    }
    lines
}

fn colorize(fg: Color, bg: Color, text: &str) -> String {
    let mut s = String::new();
    if fg != Color::Background {
        s.push_str(match fg {
            Color::Hours => "\x1b[38;2;255;167;10m",
            Color::Minutes => "\x1b[38;2;44;89;212m",
            Color::Seconds => "\x1b[38;2;80;128;80m",
            Color::Background => "",
        });
    } else {
        s.push_str("\x1b[39m");
    }
    if bg != Color::Background {
        s.push_str(match bg {
            Color::Hours => "\x1b[48;2;255;167;10m",
            Color::Minutes => "\x1b[48;2;44;89;212m",
            Color::Seconds => "\x1b[48;2;80;128;80m",
            Color::Background => "",
        });
    } else {
        s.push_str("\x1b[49m");
    }
    s.push_str(text);
    s.push_str("\x1b[0m");
    s
}
