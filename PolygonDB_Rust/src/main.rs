use serde_json::{json, Value, from_str};
use std::io::{self, BufRead};

fn main() {
    println!("Polygon v1.7 +++");

    let stdin = io::stdin();
    let mut scanner = stdin.lock();
    let mut line = String::new();
    scanner.read_line(&mut line).unwrap();

    let parsed_json: Result<Value, serde_json::Error> = from_str(&line);

    match parsed_json {
        Ok(json) => {
            println!("Parsed JSON: {:#?}", json);
        }
        Err(e) => {
            println!("Failed to parse JSON: {:?}", e);
        }
    }
}