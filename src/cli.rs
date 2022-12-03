use crate::log_level::LogLevel;
use clap::Parser;
use std::path::PathBuf;

#[derive(Parser, Debug)]
#[command(version)]
pub struct Args {
	/// DNS server host
	#[arg(long)]
	pub host: Option<String>,

	/// DNS server port
	#[arg(long)]
	pub port: Option<u16>,

	/// Log level
	#[arg(long, value_enum)]
	pub log_level: Option<LogLevel>,

	/// Config file
	#[arg(long, value_name = "FILE")]
	pub config: Option<PathBuf>,
}
