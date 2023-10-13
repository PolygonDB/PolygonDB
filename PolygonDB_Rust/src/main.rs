use json_value_remove::Remove;
use jsonptr::Pointer;
use serde::{Deserialize, Serialize};
use serde_json::Value;
use std::{io::{self, BufRead}, path::Path, fs::{self, File}};

mod maincore;
#[derive(Debug, Deserialize, Serialize)]
struct Input {
    dbname: String,
    location: String,
    action: String,
    value: serde_json::Value,
}


fn main() {
    print!("{}",execute());
}
fn execute() -> String {
    println!("Polygon v1.7 +++");


    let stdin = io::stdin();
    let mut scanner = stdin.lock();
    let mut data = String::new();
    scanner.read_line(&mut data).unwrap();

    //data = r#"{"dbname": "home", "location": "/Example", "action": "create", "value": 20}"#.to_string();
    //Example


    if is_json(&data) { //json input
        let parsed_input: Input = serde_json::from_str(&data).unwrap();
        let raw_json = fs::read_to_string(format!("databases/{}.ply", parsed_input.dbname)).expect("Unable to read file");
        let mut parsed_json: Value = serde_json::from_str(&raw_json).unwrap();
        
        if parsed_input.action == "read" {

            let output = parsed_json.pointer(&parsed_input.location);

            if output == None {
                return format!("{{\"status\": {}, \"message\": \"{:?}\"}}", 1, "None");
            } else {
                return format!("{{\"status\": {}, \"message\": \"{:?}\"}}", 0, output.unwrap());
            }

        } else if parsed_input.action == "create" {

            
            let ptr = maincore::test(&parsed_input.location, &parsed_json);
            let data_to_insert = serde_json::json!(parsed_input.value);

            let _previous = ptr.assign(&mut parsed_json, data_to_insert).unwrap();

            let json_str = serde_json::to_string_pretty(&parsed_json);

            maincore::update_content(parsed_input.dbname, json_str.unwrap().to_string());

            return format!("{{\"status\": {}, \"message\": \"{}\"}}", 0, "Successfully Created");
            


        } else if parsed_input.action == "update" { 
            
            let ptr = Pointer::try_from(parsed_input.location).unwrap();
            
            let data_to_insert = serde_json::json!(parsed_input.value);
            let _previous = ptr.assign(&mut parsed_json, data_to_insert);

            maincore::update_content(parsed_input.dbname, serde_json::to_string_pretty(&parsed_json).unwrap().to_string());

            return format!("{{\"status\": {}, \"message\": \"{}\"}}", 0, "Successfully Updated");
            
        } else if parsed_input.action == "delete" {

            let _ = parsed_json.remove(&parsed_input.location);

            maincore::update_content(parsed_input.dbname, serde_json::to_string_pretty(&parsed_json).unwrap().to_string());


            return format!("{{\"status\": {}, \"message\": \"{}\"}}", 0, "Successfully Deleted");

        } else {
            return format!("{{\"status\": {}, \"message\": \"{}\"}}", 1, "Please Pick Appropriate Command [READ/CREATE/UPDATE/DELETE]");
        }

    } else {

        let mut args: Vec<&str> = Vec::new();

        for byte in data.split_whitespace() {
            args.push(byte);
        }

        println!("{:?}", args.first());
        if  args.first().unwrap().to_string() == "CREATE_DATABASE" {
            if args.len() <= 1 {
                poly_error(0, r#"{"Error":"CREATE_DATABASE TAKES IN TWO ARGS"}"#);
                return format!("{{\"status\": {}, \"message\": \"{}\"}}", 1, "CREATE_DATABASE TAKES IN TWO ARGS");
            }

            create_database(args.get(1).unwrap().to_string());

            return format!("{{\"status\": {}, \"message\": \"{}\"}}", 0, "Successfully Created Database");
        } else {
            return format!("{{\"status\": {}, \"message\": \"{}\"}}", 0, "No Appropriate Command Selected");
        }
    }
}


fn create_database(name: String) {
    if !Path::new("databases").exists() { //Checks if the folder "databases" exists
        let _ = fs::create_dir("databases");
    }

    let _ = File::create(format!("databases/{}.ply", name));
}

fn is_json(text: &str) -> bool {
    let f = text.chars().nth(0).unwrap().to_ascii_lowercase();
    let l = text.chars().nth(text.chars().count()-1).unwrap().to_ascii_lowercase();
    if f == '{' && l == '}' {
        return true;
    }
    return false;
}

trait MyTrait {
    fn describe(&self) -> String;
}

// Implement the trait for i32 and String
impl MyTrait for i32 {
    fn describe(&self) -> String {
        format!("This is an i32: {}", self)
    }
}

impl MyTrait for String {
    fn describe(&self) -> String {
        format!("This is a String: {}", self)
    }
}


fn cleaner_output (code: i8, text: Value) -> String {
    return format!("{{\"status\": {}, \"message\": \"{}\"}}", code, text);
}

fn poly_error(erlevel: i8, text: &str){
    if erlevel == 0 { //Warning; No Real Damage Done
        print!("{}",text)
    } //Mild
    else if erlevel == 1 {} // Warning
    else if erlevel == 2 {} //Error;
    return;
}