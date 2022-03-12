use crate::{cli, log_level::LogLevel};
use anyhow::{anyhow, Context, Result};
use serde::{Deserialize, Serialize};
use std::net::{Ipv4Addr, Ipv6Addr, SocketAddr};
use trust_dns_server::client::rr::RecordType;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Config {
    #[serde(default = "default_host")]
    pub host: String,
    #[serde(default = "default_port")]
    pub port: u16,
    #[serde(default)]
    pub log_level: LogLevel,
    #[serde(default)]
    pub rule: Vec<Rule>,
}

impl Default for Config {
    fn default() -> Self {
        Config {
            host: default_host(),
            port: default_port(),
            log_level: LogLevel::Info,
            rule: vec![],
        }
    }
}

impl Config {
    pub fn from_args(args: cli::Args) -> Result<Self> {
        let mut config = Self::from_config(args.config)?;
        if let Some(host) = args.host {
            config.host = host
        }
        if let Some(port) = args.port {
            config.port = port
        }
        if let Some(log_level) = args.log_level {
            config.log_level = log_level
        }
        Ok(config)
    }

    fn from_config(config_file: Option<std::path::PathBuf>) -> Result<Self> {
        let config = match &config_file {
            None => Self::default(),
            Some(config_file) => {
                let contents = std::fs::read_to_string(config_file).with_context(|| {
                    format!("failed to read file `{}`", config_file.as_path().display())
                })?;
                serde_json::from_str(&contents).with_context(|| {
                    format!("failed to parse json `{}`", config_file.as_path().display())
                })?
            }
        };
        Ok(config)
    }
}

impl Config {
    pub fn validate_rules(&self) -> Result<()> {
        self.rule
            .iter()
            .try_for_each(|rule| self.validate_rule(rule))
    }
    fn validate_rule(&self, rule: &Rule) -> Result<()> {
        match &rule.upstream {
            Upstream::IPv4 { .. } => {
                let records = match &rule.pattern {
                    Pattern::Domain { record, .. } => record,
                    Pattern::Suffix { record, .. } => record,
                };
                match records {
                    Some(records) if records.first() == Some(&RecordType::A) => Ok(()),
                    _ => Err(anyhow!("IPv4 should be used with 'A'")),
                }
            }
            Upstream::IPv6 { .. } => {
                let records = match &rule.pattern {
                    Pattern::Domain { record, .. } => record,
                    Pattern::Suffix { record, .. } => record,
                };
                match records {
                    Some(records) if records.first() == Some(&RecordType::AAAA) => Ok(()),
                    _ => Err(anyhow!("IPv6 should be used with 'AAAA'")),
                }
            }
            _ => Ok(()),
        }
    }
}

fn default_host() -> String {
    "127.0.0.1".to_string()
}

fn default_port() -> u16 {
    0
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Rule {
    pub pattern: Pattern,
    pub upstream: Upstream,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(untagged)]
pub enum Pattern {
    Domain {
        domain: Vec<String>,
        #[serde(skip_serializing_if = "Option::is_none")]
        record: Option<Vec<RecordType>>,
    },
    Suffix {
        suffix: Vec<String>,
        #[serde(skip_serializing_if = "Option::is_none")]
        record: Option<Vec<RecordType>>,
    },
}

#[allow(clippy::upper_case_acronyms)]
#[derive(Debug, Clone, PartialEq, Eq, Hash, Serialize, Deserialize)]
#[serde(untagged)]
pub enum Upstream {
    UDP { udp: SocketAddr },
    TCP { tcp: SocketAddr },
    DoT { dot: SocketAddr, domain: String },
    DoH { doh: SocketAddr, domain: String },
    IPv4 { ipv4: Ipv4Addr },
    IPv6 { ipv6: Ipv6Addr },
    Special(SpecialUpstream),
}

#[allow(clippy::upper_case_acronyms)]
#[derive(Debug, Clone, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum SpecialUpstream {
    NXDOMAIN,
    NODATA,
}
