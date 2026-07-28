use chrono::{DateTime, Local, NaiveDateTime, TimeZone};

pub fn parse_duration(s: &str) -> Result<DateTime<Local>, String> {
    // Basic parser for "10s", "5m", "1h"
    let mut total_secs = 0;
    let mut num = 0;
    for c in s.chars() {
        if c.is_ascii_digit() {
            num = num * 10 + c.to_digit(10).unwrap() as i64;
        } else {
            match c {
                's' => total_secs += num,
                'm' => total_secs += num * 60,
                'h' => total_secs += num * 3600,
                'd' => total_secs += num * 86400,
                _ => return Err(format!("Unknown unit '{}'", c)),
            }
            num = 0;
        }
    }
    if num > 0 {
        total_secs += num; // default to seconds if no unit
    }
    
    Ok(Local::now() + chrono::Duration::seconds(total_secs))
}

pub fn parse_datetime(s: &str) -> Result<DateTime<Local>, String> {
    // Expected format: YYYY-MM-DD HH:MM:SS
    match NaiveDateTime::parse_from_str(s, "%Y-%m-%d %H:%M:%S") {
        Ok(ndt) => {
            if let chrono::LocalResult::Single(dt) = Local.from_local_datetime(&ndt) {
                Ok(dt)
            } else {
                Err("Failed to convert to local time".into())
            }
        }
        Err(e) => Err(e.to_string()),
    }
}
