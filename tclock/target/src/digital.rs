pub fn render_digits(s: &str, blink: bool) -> Vec<String> {
    let mut lines = vec![String::new(); 5];
    
    for (i, c) in s.chars().enumerate() {
        if i > 0 {
            for line in lines.iter_mut() {
                line.push(' ');
            }
        }
        let digit_idx = match c {
            '0' => 0, '1' => 1, '2' => 2, '3' => 3, '4' => 4,
            '5' => 5, '6' => 6, '7' => 7, '8' => 8, '9' => 9,
            ':' => {
                if blink { 11 } else { 10 }
            }
            _ => 11, // space or unknown
        };

        let art = get_digit_art(digit_idx);
        for (j, line) in lines.iter_mut().enumerate() {
            line.push_str(art[j]);
        }
    }
    lines
}

fn get_digit_art(idx: usize) -> [&'static str; 5] {
    match idx {
        0 => [" ━━  ", "┃  ┃ ", "     ", "┃  ┃ ", " ━━  "],
        1 => ["     ", "   ┃ ", "     ", "   ┃ ", "     "],
        2 => [" ━━  ", "   ┃ ", " ━━  ", "┃    ", " ━━  "],
        3 => [" ━━  ", "   ┃ ", " ━━  ", "   ┃ ", " ━━  "],
        4 => ["     ", "┃  ┃ ", " ━━  ", "   ┃ ", "     "],
        5 => [" ━━  ", "┃    ", " ━━  ", "   ┃ ", " ━━  "],
        6 => [" ━━  ", "┃    ", " ━━  ", "┃  ┃ ", " ━━  "],
        7 => [" ━━  ", "   ┃ ", "     ", "   ┃ ", "     "],
        8 => [" ━━  ", "┃  ┃ ", " ━━  ", "┃  ┃ ", " ━━  "],
        9 => [" ━━  ", "┃  ┃ ", " ━━  ", "   ┃ ", " ━━  "],
        10 => ["   ", "   ", ":: ", "   ", "   "],
        11 => ["   ", "   ", ".. ", "   ", "   "],
        _ => ["     ", "     ", "     ", "     ", "     "],
    }
}
