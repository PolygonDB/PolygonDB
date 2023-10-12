
use serde::{Deserialize, Serialize};
use std::{io::{self, BufRead}, path::Path, fs::{self, File}};
use colored::*;

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
    let mut data = String::new();
    scanner.read_line(&mut data).unwrap();

    let mut args: Vec<&str> = Vec::new();

    data = r#"{"dbname": "Home", "location": "", "action": "read", "value": 20}"#.to_string();


    if is_json(&data.clone()) { //json input
        
        println!("{}",data);
        let parsed_input: Input = serde_json::from_str(&data).unwrap();
        if parsed_input.action == "read" {
            println!("read");
            let parsed_json = fs::read_to_string(format!("databases/{}.ply", parsed_input.dbname)).expect("Unable to read file");
            println!("{}", parsed_json);
        }

    } else {

        for byte in data.split_whitespace() {
            args.push(byte);
        }

        println!("{:?}", args.first());
        if  args.first().unwrap().to_string() == "CREATE_DATABASE" {
            if args.len() <= 1 {
                poly_error(0, r#"{"Error":"CREATE_DATABASE TAKES IN TWO ARGS"}"#);
                return;
            }
            create_database(args.get(1).unwrap().to_string());
        }
    }
}


fn create_database(name: String) {
    if !Path::new("databases").exists() { //Checks if the folder "databases" exists
        let _ = fs::create_dir("databases");
    }

    let mut file = File::create(format!("databases/{}.ply", name));
}

fn is_json(text: &str) -> bool {
    let f = text.chars().nth(0).unwrap().to_ascii_lowercase();
    let l = text.chars().nth(text.chars().count()-1).unwrap().to_ascii_lowercase();
    if f == '{' && l == '}' {
        return true;
    }
    return false;
}

fn poly_error(erlevel: i8, text: &str){
    if erlevel == 0 { //Warning; No Real Damage Done
        print!("{}",text.bright_yellow())
    } //Mild
    else if erlevel == 1 {} // Warning
    else if erlevel == 2 {} //Error;
    return;
}