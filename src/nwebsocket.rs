use std::net::TcpListener;
use tungstenite::protocol::Message;
use std::thread::spawn;
use tungstenite::accept;

use crate::execute;

/// A WebSocket echo server
pub fn webserver (port: u16) {
    let server = TcpListener::bind(format!("0.0.0.0:{}",port)).unwrap();
    for stream in server.incoming() {
        spawn (move || {
            let mut websocket = accept(stream.unwrap()).unwrap();
            loop {
                let msg = websocket.read().unwrap();
            

                // We do not want to send back ping/pong messages.
                if msg.is_binary() || msg.is_text() {
                    let response = Message::text(execute(msg.to_string()));
                    websocket.send(response).unwrap();
                }
            }
        });
    }
}