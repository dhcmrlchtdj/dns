use crate::log_level::LogLevel;
use clap::Parser;
use std::path::PathBuf;

#[derive(Parser, Debug)]
#[command(version)]
pub struct Args {
	/// DNS server host
	#[arg(short, long)]
	pub host: Option<String>,

	/// DNS server port
	#[arg(short, long)]
	pub port: Option<u16>,

	/// Log level
	#[arg(short, long, value_enum)]
	pub log_level: Option<LogLevel>,

	/// Config file
	#[arg(short, long, value_name = "FILE")]
	pub config: Option<PathBuf>,
}
