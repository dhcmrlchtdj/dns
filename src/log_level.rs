use clap::ArgEnum;
use serde::{Deserialize, Serialize};
use std::fmt;

#[derive(Debug, Clone, ArgEnum, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum LogLevel {
    Debug, // 内部状态变化
    Info,  // 运行记录
    Warn,  // 不影响运行
    Error, // 影响运行
}

impl Default for LogLevel {
    fn default() -> Self {
        LogLevel::Info
    }
}

impl fmt::Display for LogLevel {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match self {
            LogLevel::Debug => write!(f, "debug"),
            LogLevel::Info => write!(f, "info"),
            LogLevel::Warn => write!(f, "warn"),
            LogLevel::Error => write!(f, "error"),
        }
    }
}
