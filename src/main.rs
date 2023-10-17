use mimalloc_rust::*;

#[global_allocator]
static GLOBAL_MIMALLOC: GlobalMiMalloc = GlobalMiMalloc;

use json_value_remove::Remove;
use jsonptr::Pointer;
use serde::{Deserialize, Serialize};
use serde_json::Value;
use std::{io::{self, BufRead, Write}, path::Path, fs::{self, File}, process::{self}, env};


mod maincore;
mod websocket;

#[derive(Debug, Deserialize, Serialize)]
struct Input {
    dbname: String,
    location: String,
    action: String,
    value: serde_json::Value,
}

fn main() {


    let args: Vec<String> = env::args().collect();

    //args[0] - location of .exe file
    /*
    Arguments:
    -ws -> Enable Websocket
    */
    
    if args.iter().any(|arg| arg == "-ws") {
        websocket::webserver();
    }
    
    let mut scanner = io::stdin().lock();

    loop  {
        let mut input = String::new();
        scanner.read_line(&mut input).unwrap();

        println!("{}",execute(input));

        
        io::stdout().flush().unwrap();
    }
}


pub fn execute (data: String) -> String {

    //let data = r#"{"dbname": "database", "location": "data.test", "action": "read", "value": 20}"#.to_string();
    //Examplec


    if is_json(&data) { //json input
        let parsed_input: Input = serde_json::from_str(&data).unwrap();
        if !Path::new(&format!("databases/{}.json", parsed_input.dbname)).exists() {
            return cleaner_output(1, "Database doesn't exist")
        }

        let raw_json = fs::read_to_string(format!("databases/{}.json", parsed_input.dbname)).expect("Unable to read file");
        let mut parsed_json: Value = serde_json::from_str(&raw_json).unwrap();
        
        if parsed_input.action == "read" {

            let output = parsed_json.pointer(&parsed_input.location);
            
            if output == None {
                return cleaner_output(1, "None");
            } else {
                return format!(r#"{{"Status":{}, "Message":{}}}"#, 0, output.unwrap());
            }

        } else if parsed_input.action == "create" {

            
            let ptr = maincore::define_ptr(&parsed_input.location, &parsed_json);
            let data_to_insert = serde_json::json!(parsed_input.value);

            let _previous = ptr.assign(&mut parsed_json, data_to_insert).unwrap();

            let json_str = serde_json::to_string_pretty(&parsed_json);

            maincore::update_content(parsed_input.dbname, json_str.unwrap().to_string());

            return cleaner_output(0, "Successfully CREATED json content");
            


        } else if parsed_input.action == "update" { 
            
            let ptr = Pointer::try_from(parsed_input.location).unwrap();
            
            let data_to_insert = serde_json::json!(parsed_input.value);
            let _previous = ptr.assign(&mut parsed_json, data_to_insert);

            maincore::update_content(parsed_input.dbname, serde_json::to_string_pretty(&parsed_json).unwrap().to_string());

            return cleaner_output(0, "Successfully UPDATED json content");
            
        } else if parsed_input.action == "delete" {

            let _ = parsed_json.remove(&parsed_input.location);

            maincore::update_content(parsed_input.dbname, serde_json::to_string_pretty(&parsed_json).unwrap().to_string());


            return cleaner_output(0, "Successfully DELETED json content");

        } else {
            return cleaner_output(1, "PLEASE USE THE FOLLOWING: [read/create/update/delete]");
        }

    } else {

        let mut args: Vec<&str> = Vec::new();

        for byte in data.split_whitespace() {
            args.push(byte);
        }

        if  args.first().unwrap().to_string() == "CREATE_DATABASE" {
            if args.len() <= 1 {
                return cleaner_output(1, "CREATE_DATABASE _______ <= TAKES IN ONE ARGUEMENT");
            }

            create_database(args.get(1).unwrap().to_string());

            return cleaner_output(0, "Successfully Created Database");
        } else if args.first().unwrap().to_string() == "QUIT" {
            process::exit(0);
        } else if args.first().unwrap().to_string() == "DATABASES" {


            let paths = fs::read_dir("databases").unwrap();

            let mut directory_names: Vec<String> = Vec::new();

            for path in paths {
                if let Ok(entry) = path {
                    let path_str = entry.path().display().to_string().replace("databases\\", "");
                    directory_names.push(path_str);
                }
            }

            cleaner_output(0, format!("{:?}",directory_names).as_str())

        }  else if args.first().unwrap().to_string() == "BSON" {
            if args.len() <= 1 {
                return cleaner_output(1, "CREATE_DATABASE _______ <= TAKES IN ONE ARGUEMENT");
            }

            let raw_json = fs::read_to_string(format!("databases/{}.json", args.get(1).unwrap())).expect("Unable to read file");
            //let parsed_json: Value = serde_json::from_str(&raw_json).unwrap();

            let _ = File::create(format!("databases/{}.bson",args.get(1).unwrap())).unwrap();

            fs::write(format!("databases/{}.bson",args.get(1).unwrap()), format!("{:?}", raw_json.as_bytes())).expect("Failed to create file");

            return cleaner_output(0, "Successful Conversion");

        } else if args.first().unwrap().to_string() == "BSONREAD"{
            
            if args.len() <= 1 {
                return cleaner_output(1, "CREATE_DATABASE _______ <= TAKES IN ONE ARGUEMENT");
            }

            

            let raw_json = fs::read(format!("databases/{}.bson", args.get(1).unwrap())).expect("Unable to read file");
            let parsed:Value = serde_json::from_slice(&raw_json).unwrap();

            println!("{}",parsed);

            return cleaner_output(0, "test");
        } else {
            return cleaner_output(1, "No Appropriate Function was used");
        }
    }
}

fn create_database(name: String) {
    if !Path::new("databases").exists() { //Checks if the folder "databases" exists
        let _ = fs::create_dir("databases");
    }

    let _ = File::create(format!("databases/{}.json", name));
    let _ = fs::write(format!("databases/{}.json", name), "{}");
}

fn is_json(text: &str) -> bool {
    if let Some(first) = text.chars().find(|&c| c != ' ' && c != '\n' && c != '\r') {
        if let Some(last) = text.chars().rev().find(|&c| c != ' ' && c != '\n' && c != '\r') {
            return first == '{' && last == '}';
        }
    }
    false
}


fn cleaner_output (code: i8, text: &str) -> String {
    String::from(format!("{{\"Status\":{}, \"Message\":\"{}\"}}", code, text))
}

/*fn poly_error(erlevel: i8, text: &str){
    if erlevel == 0 { //Warning; No Real Damage Done
        print!("{}",text)
    } //Mild
    else if erlevel == 1 {} // Warning
    else if erlevel == 2 {} //Error;
    return;
}*/