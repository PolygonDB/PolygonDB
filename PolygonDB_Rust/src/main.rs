use serde_json::{json, Value, from_str};
use serde::{Deserialize, Serialize};
use std::io::{self, BufRead};

#[derive(Debug, Deserialize, Serialize)]
struct Input {
    dbname: String,
    location: String,
    action: String,
    value: serde_json::Value,
}

fn main() {
    println!("Polygon v1.7 +++");

    let stdin = io::stdin();
    let mut scanner = stdin.lock();
    //let mut line = String::new();
    //scanner.read_line(&mut line).unwrap();

    let data = r#"
    {
        "dbname": "name_of_database", 
        "location": "location_in_database", 
        "action": "record", 
        "value": 20
    }"#;

    let parsed_json: Input = serde_json::from_str(data).unwrap();

    println!("{:?}", parsed_json);
}