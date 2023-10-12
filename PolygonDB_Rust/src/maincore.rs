use jsonptr::Pointer;
use serde_json::Value;
use anyhow::Result;

pub fn test(target: &str, db: &Value) -> jsonptr::Pointer{
    let ptr = Pointer::try_from(target).unwrap();
    let bar;
    if let Err(e) =  ptr.resolve(&db) {
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