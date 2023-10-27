use mimalloc_rust::*;

#[global_allocator]
static GLOBAL_MIMALLOC: GlobalMiMalloc = GlobalMiMalloc;


use lazy_static::lazy_static;

use json_value_remove::Remove;
use jsonptr::Pointer;
use serde::{Deserialize, Serialize};
use serde_json::Value;
use std::{io::{self, BufRead, Write}, path::Path, fs::{self, File}, process::{self}, env, thread};

use std::collections::HashMap;
use std::sync::Mutex;

mod maincore;
mod websocket;

lazy_static! {
    static ref MUTEX_MAP: Mutex<HashMap<String, String>> = {
        Mutex::new(HashMap::new())
    };
}

#[derive(Debug, Deserialize, Serialize)]
struct Input {
    dbname: String,
    location: String,
    action: String,
    value: serde_json::Value,
}

#[derive(Debug, Deserialize, Serialize)]
struct JsonDB {
    content: Value,
    location: String
}

impl JsonDB {
    fn init(target: String, location: String) -> JsonDB {
        // Special code goes here ...
        
        if MUTEX_MAP.lock().unwrap().get(&target).is_some() { //if exists
            let target = MUTEX_MAP.lock().unwrap().get(&target).unwrap().to_string();
            JsonDB { content: (sonic_rs::from_str(&target).unwrap()), location:location }
        } else {
            let raw_json = fs::read_to_string(format!("{}",target)).expect("Unable to read file");
            MUTEX_MAP.lock().unwrap().insert(target, raw_json.clone());
            JsonDB { content: (sonic_rs::from_str(&raw_json).unwrap()), location:location }
        }

    }

    fn read(self) -> String {
        let output = self.content.pointer(&self.location);
            
        if output == None {
            return cleaner_output(1, "None");
        } else {
            return format!(r#"{{"Status":{}, "Message":{}}}"#, 0, output.unwrap());
        }
    }

    fn create(mut self, val: Value, db: String) -> String {
            
        let ptr = maincore::define_ptr(&self.location, &self.content);
        
        let data_to_insert = serde_json::json!(val);
        let _ = ptr.assign(&mut self.content, data_to_insert).unwrap();

        thread::spawn(move || {
            maincore::update_content(db, serde_json::to_string_pretty(&self.content).unwrap().to_string());
        });

        return cleaner_output(0, "Successfully CREATED json content");
    }

    fn update(mut self, val: Value, db: String) -> String {
        let ptr = Pointer::try_from(self.location.clone()).unwrap();
            
        let data_to_insert = serde_json::json!(val);
        let _ = ptr.assign(&mut self.content, data_to_insert).unwrap();
        
        thread::spawn(move || {
            maincore::update_content(db, serde_json::to_string_pretty(&self.content).unwrap().to_string());
        });

        return cleaner_output(0, "Successfully UPDATED json content");
    }
    
    fn delete(mut self, db: String) -> String {

        
        let _ = self.content.remove(&self.location);

        thread::spawn(move || {
            maincore::update_content(db, serde_json::to_string_pretty(&self.content).unwrap().to_string());
        });  


        return cleaner_output(0, "Successfully DELETED json content");

    }
}

fn main() {
    let args: Vec<String> = env::args().skip(1).collect();

    //args[0] - location of .exe file
    /*
    Arguments:
    -ws -> Enable Websocket
    */
    
    if args.iter().any(|arg| arg == "-ws") {websocket::webserver(args[1].parse().unwrap());} // -ws 8080
    
    let mut scanner = io::stdin().lock();

    loop  {
        let mut input = String::new();
        scanner.read_line(&mut input).unwrap();

        println!("{}",execute(input));

        io::stdout().flush().unwrap();
    }
}


pub fn execute (data: String) -> String {

    if is_json(&data) { //json input
        
        let parsed_input: Input = sonic_rs::from_str(&data).unwrap();
        if !Path::new(&format!("databases/{}.json", parsed_input.dbname)).exists() {return cleaner_output(1, "Database doesn't exist")}

        let parsed_jso = JsonDB::init(format!("databases/{}.json", parsed_input.dbname), parsed_input.location);

        

        match parsed_input.action.as_str() {
            "read"=> return parsed_jso.read(),
            "create"=> return parsed_jso.create(parsed_input.value, parsed_input.dbname),
            "update"=> return parsed_jso.update(parsed_input.value, parsed_input.dbname),
            "delete"=> return parsed_jso.delete(parsed_input.dbname),
            _=> return "FAILED.".to_string()
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

            let _ = File::create(format!("databases/bin_{}.json",args.get(1).unwrap())).unwrap();

            fs::write(format!("databases/bin_{}.json",args.get(1).unwrap()), format!("{:?}", raw_json.as_bytes())).expect("Failed to create file");

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