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
                println!("A client connected with id #{}", client_id);
                // add their Responder to our `clients` map:
                clients.insert(client_id, responder);
            },
            Event::Disconnect(client_id) => {
                println!("Client #{} disconnected.", client_id);
                // remove the disconnected client from the clients map:
                clients.remove(&client_id);
            },
            Event::Message(client_id, message) => {
                //println!("Received a message from client #{}: {:?}", client_id, message);

                let mut input = String::new();
                match message.clone() {
                    Message::Text(text) => {

                        input = text;
                    }
                    Message::Binary(_) => {}
                }


                let responder = clients.get(&client_id).unwrap();

                let message: Message = Message::Text(execute(input).to_string());
                responder.send(message);
            },
        }
    }
}