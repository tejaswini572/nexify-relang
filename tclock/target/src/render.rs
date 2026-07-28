use crate::config::Config;
use chrono::{DateTime, Local};
use crossterm::{
    cursor::MoveTo,
    execute,
    style::{Color, Print, SetForegroundColor, ResetColor},
    terminal,
};
use std::io::{stdout, Write};

pub fn render_oneshot(digits: &str, _config: &Config) -> std::io::Result<()> {
    let mut lines = crate::digital::render_digits(digits, false);
    if let Some(ref color_str) = _config.color_disc {
        lines = crate::disc::apply_disc(&lines, _config, color_str);
    }
    for line in lines {
        println!("{}", line);
    }
    Ok(())
}

pub fn render_screen(config: &Config, target: Option<DateTime<Local>>, blink: bool) -> std::io::Result<()> {
    let (cols, rows) = terminal::size().unwrap_or((80, 24));
    
    // Analog bounds check
    if config.analog && (cols < 20 || rows < 10) {
        let msg = "[Terminal too small for analog]";
        execute!(stdout(), MoveTo((cols - msg.len() as u16) / 2, rows / 2))?;
        print!("{}", msg);
        return Ok(());
    }

    let time_str = get_time_string(config, target);
    
    let mut lines = if config.analog {
        let radius = std::cmp::min(cols / 2, rows) as i32 - 1;
        let time = target.unwrap_or_else(Local::now);
        crate::analog::render_analog(time, !config.no_seconds, radius)
    } else {
        crate::digital::render_digits(&time_str, if config.no_blink { false } else { blink })
    };

    if let Some(ref color_str) = config.color_disc {
        lines = crate::disc::apply_disc(&lines, config, color_str);
    }

    let box_pad = if config.box_border { 1 } else { 0 };
    let width = lines[0].chars().filter(|c| *c != '\x1b' && *c != '[' && !c.is_ascii_digit() && *c != 'm' && *c != ';').count() as u16;
    let height = lines.len() as u16;
    let total_w = width + 2 * box_pad;
    let total_h = height + 2 * box_pad;

    let mut start_x = if cols > total_w { (cols - total_w) / 2 } else { 0 };
    let mut start_y = if rows > total_h { (rows - total_h) / 2 } else { 0 };

    if config.tail.is_some() || config.digits.as_deref() == Some("-") {
        start_x = if cols > total_w { cols - total_w } else { 0 };
        start_y = 0;
    } else {
        execute!(stdout(), terminal::Clear(terminal::ClearType::All))?;
    }

    if config.box_border {
        draw_box(start_x, start_y, total_w, total_h)?;
        start_x += 1;
        start_y += 1;
    }

    let fg_color = parse_color(&config.color);

    for (i, line) in lines.iter().enumerate() {
        execute!(stdout(), MoveTo(start_x, start_y + i as u16), SetForegroundColor(fg_color), Print(line), ResetColor)?;
    }

    stdout().flush()?;
    Ok(())
}

fn draw_box(x: u16, y: u16, w: u16, h: u16) -> std::io::Result<()> {
    if w < 2 || h < 2 { return Ok(()); }
    execute!(stdout(), MoveTo(x, y), Print(format!("╭{}╮", "─".repeat((w - 2) as usize))))?;
    for i in 1..h-1 {
        execute!(stdout(), MoveTo(x, y + i), Print("│"), MoveTo(x + w - 1, y + i), Print("│"))?;
    }
    execute!(stdout(), MoveTo(x, y + h - 1), Print(format!("╰{}╯", "─".repeat((w - 2) as usize))))?;
    Ok(())
}

pub fn get_time_string(config: &Config, target: Option<DateTime<Local>>) -> String {
    let now = Local::now();
    if let Some(target_time) = target {
        let diff = target_time.signed_duration_since(now).num_seconds();
        if diff <= 0 {
            if config.no_seconds {
                return "00:00".to_string();
            } else {
                return "00:00:00".to_string();
            }
        }
        let hours = diff / 3600;
        let mins = (diff % 3600) / 60;
        let secs = diff % 60;
        if config.no_seconds {
            format!("{:02}:{:02}", hours, mins)
        } else {
            format!("{:02}:{:02}:{:02}", hours, mins, secs)
        }
    } else {
        let fmt = match (config.twenty_four_hour, config.no_seconds) {
            (true, true) => "%H:%M",
            (true, false) => "%H:%M:%S",
            (false, true) => "%I:%M",
            (false, false) => "%I:%M:%S",
        };
        now.format(fmt).to_string()
    }
}

fn parse_color(c: &str) -> Color {
    match c.to_lowercase().as_str() {
        "black" => Color::Black,
        "red" => Color::Red,
        "green" => Color::Green,
        "yellow" => Color::Yellow,
        "blue" => Color::Blue,
        "magenta" => Color::Magenta,
        "cyan" => Color::Cyan,
        "white" => Color::White,
        _ => Color::Red,
    }
}
