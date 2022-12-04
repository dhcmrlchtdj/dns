mod cli;
mod config;
mod dns_handler;
mod dns_router;
mod log_level;
mod proxy_runtime;

use anyhow::Result;
use clap::Parser;
use tracing_subscriber::{
	filter::{filter_fn, LevelFilter},
	layer::SubscriberExt,
	reload::{self, Handle},
	util::SubscriberInitExt,
	Registry,
};

use crate::config::Config;

#[tokio::main]
async fn main() -> Result<()> {
	let reload_level_filter = setup_logger();

	let args = cli::Args::parse();
	let config = Config::from_args(args)?;

	reload_level_filter.modify(|lv| *lv = config.log_level.to_tracing())?;

	config.validate_rules()?;

	let mut handler = dns_handler::DnsHandler::new();
	handler.add_rules(config.rule);

	let mut dns_server = trust_dns_server::ServerFuture::new(handler);
	let sock = tokio::net::UdpSocket::bind((config.host, config.port)).await?;
	dns_server.register_socket(sock);
	dns_server.block_until_done().await?;

	// todo, graceful shutdown
	//https://tokio.rs/tokio/topics/shutdown

	Ok(())
}

fn setup_logger() -> Handle<LevelFilter, Registry> {
	let (level_filter, reload_level_filter) = reload::Layer::new(LevelFilter::INFO);
	let target_filter = filter_fn(|l| l.target().starts_with("godns::"));
	let json_format = tracing_subscriber::fmt::layer().json();
	tracing_subscriber::registry()
		.with(level_filter)
		.with(target_filter)
		.with(json_format)
		.init();
	reload_level_filter
}
