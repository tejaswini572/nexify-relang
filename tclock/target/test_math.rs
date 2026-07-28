use std::f64::consts::PI;

fn angle_coords(max_v: f64, time_value: f64, radius: f64) -> (i32, i32) {
    let theta = 2.0 * PI * (max_v - time_value) / max_v;
    let x = (-theta.sin() * radius).round() as i32;
    let y = (-theta.cos() * radius).round() as i32;
    (x, y)
}

fn main() {
    let r = 23.0;
    // test for 14:47:43
    let sec = 43.0;
    let minute = 47.0;
    let hour = 14.0;
    let (sx, sy) = angle_coords(60.0, sec, 0.9 * r);
    let m = minute + sec / 60.0;
    let (mx, my) = angle_coords(60.0, m, 0.80 * r);
    let h = (hour % 12.0) + m / 60.0;
    let (hx, hy) = angle_coords(12.0, h, 0.47 * r);
    
    println!("Seconds: sx={}, sy={}", sx, sy);
    println!("Minutes: mx={}, my={}", mx, my);
    println!("Hours: hx={}, hy={}", hx, hy);
}
