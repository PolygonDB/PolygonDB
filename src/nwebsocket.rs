use std::net::TcpListener;
use tungstenite::protocol::Message;
use std::thread::spawn;
use tungstenite::accept;

use crate::execute;

//Experimental, needs more work on it tbh
pub fn webserver (port: u16) {
    let server = TcpListener::bind(format!("0.0.0.0:{}",port)).unwrap();
    for stream in server.incoming() {
        spawn (move || {
            let mut websocket = accept(stream.unwrap()).unwrap();
            loop {
                let mut msg = websocket.read().unwrap();
            

                // We do not want to send back ping/pong messages.
                if msg.is_binary() || msg.is_text() {
                    msg = Message::Binary(execute(msg.to_string()).into_bytes());
                    websocket.send(msg).unwrap();
                }
            }
        });
    }
}