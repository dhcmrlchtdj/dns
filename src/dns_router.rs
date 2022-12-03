use std::{collections::HashMap, sync::Arc};
use tracing::{event, Level};
use trust_dns_server::client::rr::RecordType;

use crate::config::{Pattern, Rule, Upstream};

#[derive(Debug)]
pub struct DnsRouter {
	domain: Node,
	suffix: Node,
	domain_record: HashMap<RecordType, Node>,
	suffix_record: HashMap<RecordType, Node>,
}

#[derive(Debug)]
struct Node {
	next: HashMap<String, Node>,
	matched: Option<Matched>,
}

#[derive(Debug, Clone)]
struct Matched {
	upstream: Arc<Upstream>,
	priority: usize,
}

impl DnsRouter {
	pub fn new() -> Self {
		Self {
			domain: Node::new(),
			suffix: Node::new(),
			domain_record: HashMap::new(),
			suffix_record: HashMap::new(),
		}
	}

	pub fn add_rule(&mut self, rule: Rule, priority: usize) {
		let upstream = Arc::new(rule.upstream);
		match rule.pattern {
			Pattern::Domain {
				domain,
				record: None,
			} => self.domain.add_domains(domain, upstream, priority, None),
			Pattern::Domain {
				domain,
				record: Some(record),
			} => self
				.domain_record
				.entry(record)
				.or_insert_with(Node::new)
				.add_domains(domain, upstream, priority, Some(record)),
			Pattern::Suffix {
				suffix,
				record: None,
			} => self.suffix.add_domains(suffix, upstream, priority, None),
			Pattern::Suffix {
				suffix,
				record: Some(record),
			} => self
				.suffix_record
				.entry(record)
				.or_insert_with(Node::new)
				.add_domains(suffix, upstream, priority, Some(record)),
		};
	}

	pub fn search(&self, domain: String, record: RecordType) -> Option<Arc<Upstream>> {
		event!(
			Level::TRACE,
			msg = "search",
			domain,
			record = record.to_string()
		);

		let segments = domain
			.split('.')
			.filter(|x| !x.is_empty())
			.rev()
			.enumerate()
			.collect::<Vec<(usize, &str)>>();

		let r1 = self
			.domain_record
			.get(&record)
			.and_then(|n| n.search(&segments));
		if let Some((m, len)) = r1 {
			if len == segments.len() {
				event!(
					Level::TRACE,
					msg = "found",
					domain,
					record = record.to_string(),
					upstream = serde_json::to_string(&(*m.upstream)).unwrap(),
				);
				return Some(m.upstream);
			}
		}

		let r2 = self.domain.search(&segments);
		if let Some((m, len)) = r2 {
			if len == segments.len() {
				event!(
					Level::TRACE,
					msg = "found",
					domain,
					record = record.to_string(),
					upstream = serde_json::to_string(&(*m.upstream)).unwrap(),
				);
				return Some(m.upstream);
			}
		}

		let r3 = self
			.suffix_record
			.get(&record)
			.and_then(|n| n.search(&segments));
		let r4 = self.suffix.search(&segments);
		match (r3, r4) {
			(None, None) => {
				event!(
					Level::TRACE,
					msg = "not found",
					domain,
					record = record.to_string(),
				);
				None
			}
			(Some((m, _)), None) | (None, Some((m, _))) => {
				event!(
					Level::TRACE,
					msg = "found",
					domain,
					record = record.to_string(),
					upstream = serde_json::to_string(&(*m.upstream)).unwrap(),
				);
				Some(m.upstream)
			}
			(Some((m1, _)), Some((m2, _))) => Some(if m1.priority >= m2.priority {
				event!(
					Level::TRACE,
					msg = "found",
					domain,
					record = record.to_string(),
					upstream = serde_json::to_string(&(*m1.upstream)).unwrap(),
				);
				m1.upstream
			} else {
				event!(
					Level::TRACE,
					msg = "found",
					domain,
					record = record.to_string(),
					upstream = serde_json::to_string(&(*m2.upstream)).unwrap(),
				);
				m2.upstream
			}),
		}
	}
}

impl Node {
	fn new() -> Self {
		Self {
			next: HashMap::new(),
			matched: None,
		}
	}

	fn add_domains(
		&mut self,
		domains: Vec<String>,
		upstream: Arc<Upstream>,
		priority: usize,
		record: Option<RecordType>,
	) {
		for domain in domains {
			event!(
				Level::TRACE,
				msg = "add domain",
				priority,
				domain,
				record = record.map(|x| x.to_string())
			);
			let segments = domain
				.split('.')
				.filter(|x| !x.is_empty())
				.rev()
				.collect::<Vec<&str>>();
			self.add(&segments, upstream.clone(), priority)
		}
	}

	fn add(&mut self, segments: &Vec<&str>, upstream: Arc<Upstream>, priority: usize) {
		let mut curr = self;
		for segment in segments {
			curr = curr
				.next
				.entry(segment.to_string())
				.or_insert_with(Node::new);
		}
		match curr.matched.as_ref() {
			None => curr.matched = Some(Matched::new(upstream, priority)),
			Some(m) if priority > m.priority => {
				curr.matched = Some(Matched::new(upstream, priority))
			}
			_ => (),
		};
	}

	fn search(&self, segments: &Vec<(usize, &str)>) -> Option<(Matched, usize)> {
		let mut curr = self;
		let mut matched = self.matched.clone();
		let mut longest_match = 0;
		for (idx, segment) in segments {
			match curr.next.get(*segment) {
				None => break,
				Some(next) => {
					curr = next;
					if curr.matched.is_some() {
						matched = curr.matched.clone();
						longest_match += idx + 1;
					}
				}
			};
		}
		matched.map(|m| (m, longest_match))
	}
}

impl Matched {
	fn new(upstream: Arc<Upstream>, priority: usize) -> Self {
		Self { upstream, priority }
	}
}
