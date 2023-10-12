
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

    /*data = r#"
    {
        "dbname": "name_of_database", 
        "location": "location_in_database", 
        "action": "record", 
        "value": 20
    }"#.to_string();*/


    if is_json(&data) { //json input
        let parsed_json: Input = serde_json::from_str(&data).unwrap();
        if parsed_json.action == "record" {
            println!("record");
        }
        println!("{:?}", parsed_json);

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
    let f = text.chars().nth(0).unwrap();
    let l = text.chars().nth(text.chars().count()-1).unwrap();
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