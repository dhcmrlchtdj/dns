use crate::config::{Pattern, Rule, Upstream};
use std::collections::HashMap;
use trust_dns_server::client::rr::RecordType;

#[derive(Debug, Clone)]
pub struct DnsRouter {
	domain: Node,
	suffix: Node,
	domain_record: HashMap<RecordType, Node>,
	suffix_record: HashMap<RecordType, Node>,
}

#[derive(Debug, Clone)]
struct Node {
	next: HashMap<String, Node>,
	matched: Option<Matched>,
}

#[derive(Debug, Clone)]
struct Matched {
	upstream: Upstream,
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
		let (domains, node): (Vec<String>, &Node) = match rule.pattern {
			Pattern::Domain {
				domain,
				record: None,
			} => (domain, &self.domain),
			Pattern::Domain {
				domain,
				record: Some(record),
			} => (
				domain,
				self.domain_record.entry(record).or_insert_with(Node::new),
			),
			Pattern::Suffix {
				suffix,
				record: None,
			} => (suffix, &self.suffix),
			Pattern::Suffix {
				suffix,
				record: Some(record),
			} => (
				suffix,
				self.suffix_record.entry(record).or_insert_with(Node::new),
			),
		};

		domains.into_iter().for_each(|domain| {
			let segments = domain
				.split('.')
				.filter(|x| !x.is_empty())
				.rev()
				.collect::<Vec<&str>>();
			node.add(segments, rule.upstream.clone(), priority)
		});
	}

	pub fn search(&self, domain: String, record_type: RecordType) -> Option<Upstream> {
		let segments = domain
			.split('.')
			.filter(|x| !x.is_empty())
			.collect::<Vec<&str>>();

		if let Some((m, len)) = self
			.domain_record
			.get(&record_type)
			.and_then(|n| n.search(&segments))
		{
			if len == segments.len() {
				return Some(m.upstream);
			}
		}

		if let Some((m, len)) = self.domain.search(&segments) {
			if len == segments.len() {
				return Some(m.upstream);
			}
		}

		let r1 = self
			.suffix_record
			.get(&record_type)
			.and_then(|n| n.search(&segments));
		let r2 = self.suffix.search(&segments);
		match (r1, r2) {
			(None, None) => None,
			(Some((m, _)), None) | (None, Some((m, _))) => Some(m.upstream),
			(Some((m1, _)), Some((m2, _))) => Some(if m1.priority >= m2.priority {
				m1.upstream
			} else {
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

	fn add(&mut self, mut segments: Vec<&str>, upstream: Upstream, priority: usize) {
		let mut curr = self;
		for segment in segments {
			curr = curr
				.next
				.entry(segment.to_string())
				.or_insert_with(Node::new);
		}
		match curr.matched.as_ref() {
			None => self.matched = Some(Matched::new(upstream, priority)),
			Some(m) if priority > m.priority => {
				self.matched = Some(Matched::new(upstream, priority))
			}
			_ => (),
		};
	}

	fn search(&self, mut segments: &Vec<&str>) -> Option<(Matched, usize)> {
		let mut curr = self;
		let mut matched = self.matched;
		let mut longestMatch = 0;
		for (idx, segment) in segments.iter().enumerate() {
			match curr.next.get(*segment) {
				None => break,
				Some(next) => {
					match (matched, next.matched) {
						(None, Some(m)) => {
							matched = next.matched;
							longestMatch = idx + 1;
						}
						(Some(m1), Some(m2)) if m2.priority > m1.priority => {
							matched = next.matched;
							longestMatch = idx + 1;
						}
						_ => (),
					}
					curr = next;
				}
			};
		}
		return matched.and_then(|m| Some((m, longestMatch)));
	}
}

impl Matched {
	fn new(upstream: Upstream, priority: usize) -> Self {
		Self { upstream, priority }
	}
}
