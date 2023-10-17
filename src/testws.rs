use fastwebsockets::{FragmentCollector, WebSocket, Role};
use tokio::net::TcpStream;
use anyhow::Result;


async fn handle(
    socket: TcpStream,
  ) -> Result<()> {
    let mut ws = WebSocket::after_handshake(socket, Role::Server);
    let mut ws = FragmentCollector::new(ws);
    let incoming = ws.read_frame().await?;
    // Always returns full messages
    assert!(incoming.fin);
    Ok(())
  }