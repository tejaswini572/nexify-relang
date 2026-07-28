use std::env;
use std::io::{self, Read, Write};

mod tables;
mod qr;
mod matrix;
mod render;
mod reed_solomon;
mod masking;
mod encoding;

fn main() {
    let mut verbose = false;
    let mut level = "L".to_string();
    let mut quiet_zone = 2;
    let mut input = String::new();
    let mut args: Vec<String> = env::args().collect();
    args.remove(0);

    let mut i = 0;
    while i < args.len() {
        match args[i].as_str() {
            "-v" => {
                verbose = true;
                i += 1;
            }
            "-l" => {
                if i + 1 < args.len() {
                    level = args[i + 1].clone();
                    i += 2;
                } else {
                    i += 1;
                }
            }
            "-q" => {
                if i + 1 < args.len() {
                    quiet_zone = args[i + 1].parse().unwrap_or(2);
                    i += 2;
                } else {
                    i += 1;
                }
            }
            "-s" => {
                // ignore sixel flag
                i += 1;
            }
            arg => {
                if input.is_empty() {
                    input.push_str(arg);
                } else {
                    input.push(' ');
                    input.push_str(arg);
                }
                i += 1;
            }
        }
    }

    if input.is_empty() {
        let mut buf = String::new();
        let _ = io::stdin().read_to_string(&mut buf);
        input = buf;
    }

    let ec_level = match level.to_lowercase().as_str() {
        "l" => qr::Level::L,
        "m" => qr::Level::M,
        "q" => qr::Level::Q,
        "h" => qr::Level::H,
        _ => {
            eprintln!("Invalid error correction level: {}", level);
            eprintln!("Valid options are [L, M, Q, H]");
            std::process::exit(1);
        }
    };

    if verbose {
        println!("Level: {}", level);
        println!("Quietzone Border Size: {}", quiet_zone);
        println!("Encoded data: {}", input);
        println!();
    }

    let code = qr::encode(&input, ec_level);
    println!();
    crate::render::write_full_blocks(&mut std::io::stdout(), &code, quiet_zone);
}
