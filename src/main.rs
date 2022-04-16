mod cli;
mod config;
mod dns_handler;
mod dns_router;
mod log_level;
mod proxy_runtime;

use anyhow::Result;
use clap::Parser;
use tracing::Level;

use crate::config::Config;

#[tokio::main]
async fn main() -> Result<()> {
    let args = cli::Args::parse();
    let config = Config::from_args(args)?;

    tracing_subscriber::fmt::fmt()
        .with_max_level(Level::DEBUG)
        .init();

    config.validate_rules()?;

    let mut handler = dns_handler::DnsHandler::new();
    handler.add_rules(config.rule);

    let mut dns_server = trust_dns_server::ServerFuture::new(handler);
    let sock = tokio::net::UdpSocket::bind((config.host, config.port)).await?;
    dns_server.register_socket(sock);
    dns_server.block_until_done().await?;

    Ok(())
}
