use crate::log_level::LogLevel;
use clap::Parser;
use std::path::PathBuf;

// document for clap
// https://github.com/clap-rs/clap/blob/v3.1.6/examples/derive_ref/README.md

#[derive(Parser, Debug)]
#[clap(version)]
pub struct Args {
    /// DNS server host
    #[clap(short, long)]
    pub host: Option<String>,

    /// DNS server port
    #[clap(short, long)]
    pub port: Option<u16>,

    /// Log level
    #[clap(short, long, arg_enum)]
    pub log_level: Option<LogLevel>,

    /// Config file
    #[clap(short, long, parse(from_os_str), value_name = "FILE")]
    pub config: Option<PathBuf>,
}
