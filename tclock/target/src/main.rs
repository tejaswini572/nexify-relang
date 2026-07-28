mod config;
mod analog;
mod digital;
mod tail;
mod countdown;
mod render;
mod disc;

use config::Config;
use crossterm::{
    cursor::{Hide, Show},
    event::{self, Event, KeyCode, KeyModifiers},
    execute,
    terminal::{disable_raw_mode, enable_raw_mode, EnterAlternateScreen, LeaveAlternateScreen},
};
use std::io::{self, stdout};
use std::process;
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc;
use std::panic;
use std::time::Duration;

pub fn cleanup_terminal() {
    let _ = disable_raw_mode();
    let _ = execute!(stdout(), Show, LeaveAlternateScreen);
}

fn setup_panic_hook() {
    let original_hook = panic::take_hook();
    panic::set_hook(Box::new(move |panic_info| {
        cleanup_terminal();
        original_hook(panic_info);
    }));
}

fn main() -> io::Result<()> {
    let config = Config::parse();

    if let Some(ref digits) = config.digits {
        if digits != "-" {
            render::render_oneshot(digits, &config)?;
            return Ok(());
        }
    }

    let target_time = if let Some(ref cd_str) = config.countdown {
        Some(countdown::parse_duration(cd_str).unwrap_or_else(|e| {
            eprintln!("Invalid countdown duration: {}", e);
            process::exit(1);
        }))
    } else if let Some(ref until_str) = config.until {
        Some(countdown::parse_datetime(until_str).unwrap_or_else(|e| {
            eprintln!("Invalid until datetime: {}", e);
            process::exit(1);
        }))
    } else {
        None
    };

    let is_tail = config.tail.is_some() || config.digits.as_deref() == Some("-");
    if is_tail {
        setup_panic_hook();
        let sig_handled = Arc::new(AtomicBool::new(false));
        let sig_clone = sig_handled.clone();
        ctrlc::set_handler(move || {
            sig_clone.store(true, Ordering::SeqCst);
        }).expect("Error setting Ctrl-C handler");
        return tail::run_tail_loop(&config, target_time, sig_handled);
    }

    setup_panic_hook();
    let sig_handled = Arc::new(AtomicBool::new(false));
    let sig_clone = sig_handled.clone();
    ctrlc::set_handler(move || {
        sig_clone.store(true, Ordering::SeqCst);
    }).expect("Error setting Ctrl-C handler");

    enable_raw_mode()?;
    execute!(stdout(), EnterAlternateScreen, Hide)?;

    let mut blink = false;
    let mut last_second = chrono::Local::now().timestamp();

    loop {
        if sig_handled.load(Ordering::SeqCst) {
            break;
        }

        render::render_screen(&config, target_time, blink)?;

        if event::poll(Duration::from_millis(100))? {
            match event::read()? {
                Event::Key(key_event) => {
                    match key_event.code {
                        KeyCode::Char('q') | KeyCode::Esc => break,
                        KeyCode::Char('c') if key_event.modifiers.contains(KeyModifiers::CONTROL) => break,
                        _ => {}
                    }
                }
                Event::Resize(_, _) => {
                    execute!(stdout(), crossterm::terminal::Clear(crossterm::terminal::ClearType::All))?;
                }
                _ => {}
            }
        }

        let current_second = chrono::Local::now().timestamp();
        if current_second != last_second {
            last_second = current_second;
            blink = !blink;
        }
    }

    cleanup_terminal();
    Ok(())
}
