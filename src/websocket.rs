/*
Websocket.rs

This is the websocket code
*/
use simple_websockets::{Event, Responder, Message};
use std::collections::HashMap;

use crate::execute;

pub fn webserver(port: u16) {
    // listen for WebSockets on port 8080:
    let event_hub = simple_websockets::launch(port)
        .expect("failed to listen on port");

    println!("Server is launced on port {}",port);
    let mut clients: HashMap<u64, Responder> = HashMap::new();
    

    loop {
        match event_hub.poll_event() {
            Event::Connect(client_id, responder) => {
                clients.insert(client_id, responder);
            },
            Event::Disconnect(client_id) => {
                clients.remove(&client_id);
            },
            Event::Message(client_id, message) => {

                //let mut input = String::new();
                match message {
                    Message::Text(text) => {

                        let responder = clients.get(&client_id).unwrap();

                        let message: Message = Message::Binary(execute(text).into_bytes());
                        responder.send(message);                    
                    }
                    Message::Binary(_) => {}
                }


                //let responder = clients.get(&client_id).unwrap();

                //let message: Message = Message::Binary(execute(input).into_bytes());
                //responder.send(message);
            },
        }
    }
}