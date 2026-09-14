use std::io::{self, BufRead, Write};

use chess::{UciAction, UciSession};

fn main() {
    let stdin = io::stdin();
    let mut stdout = io::stdout();
    let mut session = UciSession::new();

    for line in stdin.lock().lines() {
        let Ok(line) = line else {
            break;
        };
        match session.handle_line(&line) {
            UciAction::Reply(replies) => {
                for reply in replies {
                    let _ = writeln!(stdout, "{reply}");
                }
                let _ = stdout.flush();
            }
            UciAction::Quit => break,
        }
    }
}
