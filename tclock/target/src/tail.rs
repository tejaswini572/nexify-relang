use crate::config::Config;
use chrono::{DateTime, Local};
use crossterm::{
    cursor::{RestorePosition, SavePosition},
    execute,
};
use std::io::{self, stdout, Read, Write};
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc;
use std::time::Duration;
use std::fs::File;

pub fn run_tail_loop(
    config: &Config,
    target: Option<DateTime<Local>>,
    sig_handled: Arc<AtomicBool>,
) -> io::Result<()> {
    let mut is_stdin = false;
    let mut file: Option<File> = None;
    if let Some(ref t) = config.tail {
        if t == "-" {
            is_stdin = true;
        } else {
            file = Some(File::open(t)?);
        }
    } else if config.digits.as_deref() == Some("-") {
        is_stdin = true;
    }

    let mut blink = false;
    let mut last_second = chrono::Local::now().timestamp();
    
    let (tx, rx) = std::sync::mpsc::channel();
    if is_stdin {
        std::thread::spawn(move || {
            let mut buf = [0; 1024];
            let mut stdin = io::stdin();
            loop {
                if let Ok(n) = stdin.read(&mut buf) {
                    if n == 0 {
                        std::thread::sleep(Duration::from_millis(100));
                    } else {
                        if tx.send(buf[..n].to_vec()).is_err() { break; }
                    }
                } else {
                    std::thread::sleep(Duration::from_millis(100));
                }
            }
        });
    } else if let Some(mut f) = file {
        std::thread::spawn(move || {
            let mut buf = [0; 1024];
            loop {
                if let Ok(n) = f.read(&mut buf) {
                    if n == 0 {
                        std::thread::sleep(Duration::from_millis(100));
                    } else {
                        if tx.send(buf[..n].to_vec()).is_err() { break; }
                    }
                } else {
                    std::thread::sleep(Duration::from_millis(100));
                }
            }
        });
    }

    loop {
        if sig_handled.load(Ordering::SeqCst) {
            break;
        }

        let mut output_occurred = false;
        while let Ok(data) = rx.try_recv() {
            stdout().write_all(&data)?;
            output_occurred = true;
        }

        let current_second = chrono::Local::now().timestamp();
        if current_second != last_second {
            last_second = current_second;
            blink = !blink;
            output_occurred = true;
        }

        if output_occurred {
            execute!(stdout(), SavePosition)?;
            let mut modified_config = config.clone();
            modified_config.box_border = true;
            
            crate::render::render_screen(&modified_config, target, blink)?;
            
            execute!(stdout(), RestorePosition)?;
            stdout().flush()?;
        }

        std::thread::sleep(Duration::from_millis(50));
    }
    
    Ok(())
}
