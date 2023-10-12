#![allow(dead_code)]
use jsonptr::Pointer;
use serde::{Deserialize, Serialize};
use serde_json::Value;
use std::{io::{self, BufRead, BufWriter}, path::Path, fs::{self, File}, str::FromStr};

mod maincore;
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

    data = r#"{"dbname": "home", "location": "/Example", "action": "update", "value": 20}"#.to_string();
    //Example


    if is_json(&data.clone()) { //json input
        let parsed_input: Input = serde_json::from_str(&data).unwrap();
        let raw_json = fs::read_to_string(format!("databases/{}.ply", parsed_input.dbname)).expect("Unable to read file");
        let mut parsed_json: Value = serde_json::from_str(&raw_json).unwrap();
        
        if parsed_input.action == "read" {

            let output = parsed_json.pointer(&parsed_input.location);

            if output == None {
                println!("None");
            } else {
                println!("{:?}", output.unwrap());
            }

        } else if (parsed_input.action == "create") {

            
            let ptr = maincore::test(&parsed_input.location, &parsed_json);
            let data_to_insert = serde_json::json!(parsed_input.value);

            let _previous = ptr.assign(&mut parsed_json, data_to_insert).unwrap();


            println!("{:?}",parsed_json);


        } else if (parsed_input.action == "update") { 
            let ptr = Pointer::try_from(parsed_input.location).unwrap();
            
            let data_to_insert = serde_json::json!(parsed_input.value);
            let _previous = ptr.assign(&mut parsed_json, data_to_insert).unwrap();

            println!("{:?}",parsed_json)
            
        } else if (parsed_input.action == "delete") {
            let ptr = Pointer::try_from(parsed_input.location).unwrap();
            let _previous = ptr.delete(&mut parsed_json).unwrap();
        } else {
            poly_error(0, "NO ACTION WAS PICKED. [read/create/update/delete]");
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
        print!("{}",text)
    } //Mild
    else if erlevel == 1 {} // Warning
    else if erlevel == 2 {} //Error;
    return;
}