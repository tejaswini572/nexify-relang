use crate::config::Config;

fn parse_hex_color(hex: &str) -> Option<(u8, u8, u8)> {
    let hex = hex.trim_start_matches('#');
    if hex.len() == 6 {
        let r = u8::from_str_radix(&hex[0..2], 16).ok()?;
        let g = u8::from_str_radix(&hex[2..4], 16).ok()?;
        let b = u8::from_str_radix(&hex[4..6], 16).ok()?;
        Some((r, g, b))
    } else {
        match hex.to_lowercase().as_str() {
            "black" => Some((0, 0, 0)),
            "red" => Some((255, 0, 0)),
            "green" => Some((0, 255, 0)),
            "yellow" => Some((255, 255, 0)),
            "blue" => Some((0, 0, 255)),
            "magenta" => Some((255, 0, 255)),
            "cyan" => Some((0, 255, 255)),
            "white" => Some((255, 255, 255)),
            _ => None,
        }
    }
}

fn get_color(x: f32, y: f32, cx: f32, cy: f32, r: f32, aliasing: f32, disc_color: (u8, u8, u8)) -> Option<(u8, u8, u8)> {
    let dx = x - cx;
    let dy = (y - cy) * 2.0; // Terminal characters are ~twice as tall as they are wide
    let dist = (dx * dx + dy * dy).sqrt();

    // Use a fixed thin band for aliasing, scaled slightly by the aliasing factor.
    // 0 = sharp edge, 1 = softer rim.
    let aliasing_band = aliasing * 6.0; 
    let edge_min = (r - aliasing_band).max(0.0);
    let edge_max = r;
    
    let factor = if dist <= edge_min {
        1.0
    } else if dist >= edge_max {
        0.0
    } else {
        // smoothstep blend
        let t = (edge_max - dist) / (edge_max - edge_min);
        t * t * (3.0 - 2.0 * t)
    };

    if factor <= 0.0 {
        None
    } else {
        Some((
            (disc_color.0 as f32 * factor) as u8,
            (disc_color.1 as f32 * factor) as u8,
            (disc_color.2 as f32 * factor) as u8,
        ))
    }
}

fn rgb_to_ansi256(r: u8, g: u8, b: u8) -> u8 {
    if r == g && g == b {
        if r < 8 { 16 } else if r > 248 { 231 } else {
            (((r as f32 - 8.0) / 247.0) * 24.0).round() as u8 + 232
        }
    } else {
        let ri = (r as f32 / 255.0 * 5.0).round() as u8;
        let gi = (g as f32 / 255.0 * 5.0).round() as u8;
        let bi = (b as f32 / 255.0 * 5.0).round() as u8;
        16 + 36 * ri + 6 * gi + bi
    }
}

fn format_ansi(is_bg: bool, color: (u8, u8, u8), truecolor: bool) -> String {
    if truecolor {
        format!("\x1b[{};2;{};{};{}m", if is_bg { 48 } else { 38 }, color.0, color.1, color.2)
    } else {
        format!("\x1b[{};5;{}m", if is_bg { 48 } else { 38 }, rgb_to_ansi256(color.0, color.1, color.2))
    }
}

pub fn apply_disc(lines: &[String], config: &Config, color_str: &str) -> Vec<String> {
    let disc_color = parse_hex_color(color_str).unwrap_or((224, 192, 32)); // E0C020 default
    let fg_color = parse_hex_color(&config.color).unwrap_or((255, 0, 0)); // default red
    
    // Check if TrueColor is supported
    let truecolor = std::env::var("COLORTERM").map(|v| v == "truecolor" || v == "24bit").unwrap_or(false)
        && std::env::var("TERM_PROGRAM").map(|v| v != "Apple_Terminal").unwrap_or(true);
    
    let h = lines.len();
    let w = if h > 0 { lines[0].chars().count() } else { 0 };

    let cx = w as f32 / 2.0;
    let cy = h as f32 / 2.0;
    let r = w as f32 * config.radius;

    let max_dist_x = r;
    let max_dist_y = r / 2.0;
    
    // Calculate bounding box that contains both the text and the disc
    let min_x = ((cx - max_dist_x - 1.0).floor() as isize).min(0);
    let max_x = ((cx + max_dist_x + 1.0).ceil() as isize).max(w as isize - 1);
    let min_y = ((cy - max_dist_y - 1.0).floor() as isize).min(0);
    let max_y = ((cy + max_dist_y + 1.0).ceil() as isize).max(h as isize - 1);

    let get_bg = |x_pos: f32, y_pos: f32| -> Option<(u8, u8, u8)> {
        let pad_x = 2.0;
        let pad_y = 1.0;
        if x_pos >= -pad_x && x_pos < (w as f32) + pad_x && y_pos >= -pad_y && y_pos < (h as f32) + pad_y {
            Some((0, 0, 0))
        } else {
            get_color(x_pos, y_pos, cx, cy, r, config.aliasing, disc_color)
        }
    };

    let mut out_lines = Vec::new();
    for y in min_y..=max_y {
        let mut line_str = String::new();
        for x in min_x..=max_x {
            let ch = if x >= 0 && x < w as isize && y >= 0 && y < h as isize {
                lines[y as usize].chars().nth(x as usize).unwrap_or(' ')
            } else {
                ' '
            };

            if ch != ' ' {
                // Text cell - always inside the backdrop
                line_str.push_str(&format_ansi(false, fg_color, truecolor));
                line_str.push_str(&format_ansi(true, (0, 0, 0), truecolor)); // black backdrop
                line_str.push(ch);
            } else {
                // Empty cell, use half blocks
                let top_bg = get_bg(x as f32 + 0.5, y as f32 + 0.25);
                let bot_bg = get_bg(x as f32 + 0.5, y as f32 + 0.75);

                let top_c = top_bg.unwrap_or(if config.black_bg { (0,0,0) } else { (0,0,0) });
                let bot_c = bot_bg.unwrap_or(if config.black_bg { (0,0,0) } else { (0,0,0) });

                let top_transparent = top_bg.is_none() && !config.black_bg;
                let bot_transparent = bot_bg.is_none() && !config.black_bg;

                if top_transparent && bot_transparent {
                    line_str.push_str("\x1b[0m "); // reset and space
                } else if top_transparent {
                    // top is transparent, bottom is colored
                    line_str.push_str(&format!("\x1b[49m{}▄", format_ansi(false, bot_c, truecolor)));
                } else if bot_transparent {
                    // top is colored, bottom is transparent
                    line_str.push_str(&format!("\x1b[49m{}▀", format_ansi(false, top_c, truecolor)));
                } else {
                    // both colored
                    line_str.push_str(&format!("{}{}▀", format_ansi(true, bot_c, truecolor), format_ansi(false, top_c, truecolor)));
                }
            }
        }
        line_str.push_str("\x1b[0m"); // reset at end of line
        out_lines.push(line_str);
    }
    out_lines
}
