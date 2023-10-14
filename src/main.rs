#[global_allocator]
static GLOBAL: mimalloc::MiMalloc = mimalloc::MiMalloc;

use json_value_remove::Remove;
use jsonptr::Pointer;
use serde::{Deserialize, Serialize};
use serde_json::Value;
use std::{io::{self, BufRead, Write}, path::Path, fs::{self, File}, process::{self}, env};

mod maincore;
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
    -o -> Output the context to text file 
    */

    let to_text = args.iter().any(|arg| arg == "-o");

    loop  {
        static mut data: String = String::new();

        if to_text {
            let input = execute(data);
            let mut file = File::create("output.txt").unwrap();
            file.write(format!("{}",input).as_bytes() ).expect("write failed");
        } else {
            println!("{}",execute(data));
        }
        
        io::stdout().flush().unwrap();
    }

}
fn execute(mut data: String) -> String {


    let stdin = io::stdin();
    let mut scanner = stdin.lock();
    scanner.read_line(&mut data).unwrap();



    //data = r#"{"dbname": "database", "location": "/data", "action": "read", "value": 20}"#.to_string();
    //Example


    if is_json(&data) { //json input
        let parsed_input: Input = serde_json::from_str(&data).unwrap();
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

            
            let ptr = maincore::test(&parsed_input.location, &parsed_json);
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
    let temp = text.replace("\n", "").replace("\r", "");

    let l = temp.chars().rev().nth(0).unwrap();
    let f = temp.chars().nth(0).unwrap();

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


fn cleaner_output (code: i8, text: &str) -> String {
    return format!(r#"{{"Status":{}, "Message":"{}"}}"#, code, text);
}

/*fn poly_error(erlevel: i8, text: &str){
    if erlevel == 0 { //Warning; No Real Damage Done
        print!("{}",text)
    } //Mild
    else if erlevel == 1 {} // Warning
    else if erlevel == 2 {} //Error;
    return;
}*/