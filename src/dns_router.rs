use crate::config::{Pattern, Rule, Upstream};
use std::collections::HashMap;
use trust_dns_server::client::rr::RecordType;

#[derive(Debug, Clone)]
pub struct DnsRouter {
	record_router: HashMap<RecordType, Node>,
	default_router: Node,
}

impl DnsRouter {
	pub fn new() -> Self {
		Self {
			record_router: HashMap::new(),
			default_router: Node::new(),
		}
	}

	pub fn add_rule(&mut self, rule: Rule, index: usize) {
		let (is_suffix, domains, record) = match rule.pattern {
			Pattern::Domain { domain, record } => (false, domain, record),
			Pattern::Suffix { suffix, record } => (true, suffix, record),
		};
		domains.into_iter().for_each(|domain| {
			let segments = domain
				.split('.')
				.filter(|x| !x.is_empty())
				.rev()
				.collect::<Vec<&str>>();
			match record {
				None => self
					.default_router
					.add(is_suffix, segments, rule.upstream.clone(), index),
				Some(record) => {
					let node = self.record_router.entry(record).or_insert_with(Node::new);
					node.add(is_suffix, segments.clone(), rule.upstream.clone(), index);
				}
			};
		});
	}

	pub fn search(&self, domain: String, record_type: RecordType) -> Option<Upstream> {
		let segments = domain
			.split('.')
			.filter(|x| !x.is_empty())
			.collect::<Vec<&str>>();

		let record_upstream: Option<(Upstream, bool, usize)> =
			if let Some(r) = self.record_router.get(&record_type) {
				r.search(segments.clone())
			} else {
				None
			};
		let default_upstream = self.default_router.search(segments);

		match (record_upstream, default_upstream) {
			(None, None) => None,
			(Some((u, _, _)), None) | (None, Some((u, _, _))) => Some(u),
			(Some((u, false, _)), Some(_)) => Some(u),
			(Some(_), Some((u, false, _))) => Some(u),
			(Some((u1, true, i1)), Some((u2, true, i2))) => Some(if i1 < i2 { u1 } else { u2 }),
		}
	}
}

#[derive(Debug, Clone)]
struct Node {
	next: HashMap<String, Node>,
	domain: Option<(Upstream, bool, usize)>, // upstream, is_suffix, index
	suffix: Option<(Upstream, bool, usize)>,
}

impl Node {
	fn new() -> Self {
		Self {
			next: HashMap::new(),
			domain: None,
			suffix: None,
		}
	}

	fn add(&mut self, is_suffix: bool, mut segments: Vec<&str>, upstream: Upstream, index: usize) {
		if segments.is_empty() {
			match (is_suffix, self.suffix.as_ref(), self.domain.as_ref()) {
				(true, None, _) => self.suffix = Some((upstream, is_suffix, index)),
				(true, Some((_, _, curr)), _) if index < *curr => {
					self.suffix = Some((upstream, is_suffix, index))
				}
				(false, _, None) => self.domain = Some((upstream, is_suffix, index)),
				(false, _, Some((_, _, curr))) if index < *curr => {
					self.domain = Some((upstream, is_suffix, index))
				}
				_ => (),
			};
		} else {
			let segment = segments.pop().unwrap();
			let next = self
				.next
				.entry(segment.to_string())
				.or_insert_with(Node::new);
			next.add(is_suffix, segments, upstream, index);
		}
	}

	fn search(&self, mut segments: Vec<&str>) -> Option<(Upstream, bool, usize)> {
		if segments.is_empty() {
			match (self.domain.as_ref(), self.suffix.as_ref()) {
				(Some(m), _) => Some(m.clone()),
				(None, Some(m)) => Some(m.clone()),
				(None, None) => None,
			}
		} else {
			let segment = segments.pop().unwrap();
			match self.next.get(segment) {
				None => self.suffix.clone(),
				Some(next) => match (next.search(segments), self.suffix.as_ref()) {
					(Some(m), _) => Some(m),
					(None, Some(m)) => Some(m.clone()),
					(None, None) => None,
				},
			}
		}
	}
}
