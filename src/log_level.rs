use clap::ArgEnum;
use serde::{Deserialize, Serialize};
use std::fmt;

#[derive(Debug, Clone, ArgEnum, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum LogLevel {
	Trace,
	Debug,
	Info,
	Warn,
	Error,
}

impl Default for LogLevel {
	fn default() -> Self {
		LogLevel::Info
	}
}

impl fmt::Display for LogLevel {
	fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
		match self {
			LogLevel::Trace => write!(f, "trace"),
			LogLevel::Debug => write!(f, "debug"),
			LogLevel::Info => write!(f, "info"),
			LogLevel::Warn => write!(f, "warn"),
			LogLevel::Error => write!(f, "error"),
		}
	}
}
