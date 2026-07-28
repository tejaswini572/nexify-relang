use std::env;

#[derive(Debug, Clone)]
pub struct Config {
    pub digits: Option<String>,
    pub analog: bool,
    pub twenty_four_hour: bool,
    pub no_seconds: bool,
    pub no_blink: bool,
    pub box_border: bool,
    pub color: String,
    pub countdown: Option<String>,
    pub until: Option<String>,
    pub tail: Option<String>,
    pub color_disc: Option<String>,
    pub radius: f32,
    pub aliasing: f32,
    pub black_bg: bool,
}

impl Config {
    pub fn parse() -> Self {
        let mut config = Config {
            digits: None,
            analog: false,
            twenty_four_hour: false,
            no_seconds: false,
            no_blink: false,
            box_border: false,
            color: "red".to_string(),
            countdown: None,
            until: None,
            tail: None,
            color_disc: None,
            radius: 1.2,
            aliasing: 0.8,
            black_bg: false,
        };

        let mut args = env::args().skip(1).peekable();
        while let Some(arg) = args.next() {
            match arg.as_str() {
                "-analog" => config.analog = true,
                "-24" => config.twenty_four_hour = true,
                "-no-seconds" => config.no_seconds = true,
                "-no-blink" => config.no_blink = true,
                "-box" => config.box_border = true,
                "-color" => {
                    if let Some(c) = args.next() { config.color = c; }
                    else { eprintln!("Error: -color requires a value"); std::process::exit(1); }
                }
                "-countdown" => {
                    if let Some(cd) = args.next() { config.countdown = Some(cd); }
                    else { eprintln!("Error: -countdown requires a value"); std::process::exit(1); }
                }
                "-until" => {
                    if let Some(u) = args.next() { config.until = Some(u); }
                    else { eprintln!("Error: -until requires a value"); std::process::exit(1); }
                }
                "-tail" => {
                    if let Some(t) = args.next() { config.tail = Some(t); }
                    else { eprintln!("Error: -tail requires a value"); std::process::exit(1); }
                }
                "-color-disc" => {
                    if let Some(next) = args.peek() {
                        if !next.starts_with('-') {
                            config.color_disc = Some(args.next().unwrap());
                        } else {
                            config.color_disc = Some("E0C020".to_string());
                        }
                    } else {
                        config.color_disc = Some("E0C020".to_string());
                    }
                }
                "-radius" => {
                    if let Some(r) = args.next() {
                        config.radius = r.parse().unwrap_or(1.2);
                    } else {
                        eprintln!("Error: -radius requires a value"); std::process::exit(1);
                    }
                }
                "-aliasing" => {
                    if let Some(a) = args.next() {
                        config.aliasing = a.parse().unwrap_or(0.8);
                    } else {
                        eprintln!("Error: -aliasing requires a value"); std::process::exit(1);
                    }
                }
                "-black-bg" => config.black_bg = true,
                _ => {
                    if arg.starts_with('-') && arg != "-" {
                        eprintln!("Unknown flag: {}", arg);
                        std::process::exit(1);
                    }
                    if config.digits.is_none() {
                        config.digits = Some(arg);
                    }
                }
            }
        }
        config
    }
}
