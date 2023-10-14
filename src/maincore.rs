/*
maincore.rs

This basically utility code
*/
use std::fs;

use jsonptr::Pointer;
use serde_json::Value;

pub fn define_ptr(target: &str, db: &Value) -> jsonptr::Pointer{
    let ptr = Pointer::try_from(target).unwrap();
    let bar;
    if let Err(_e) =  ptr.resolve(&db) {
        return Pointer::try_from(target).unwrap(); //if failure, chances are is that it doesn't exist
    } else {
        bar = ptr.resolve(&db).unwrap();
    }

    if is_array(bar) {
        let t = bar.as_array().unwrap().len();
        return Pointer::try_from(format!("{}/{}",target, t)).unwrap()
    } else {
        return Pointer::try_from(target).unwrap()
    }
}

fn is_array(value: &Value) -> bool {
    value.is_array()
}

pub fn update_content(dbname: String, content: String) -> bool {
    let _ = fs::write(format!("databases/{}.json",dbname), content);
    return false;
}