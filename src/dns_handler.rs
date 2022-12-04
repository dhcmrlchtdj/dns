use async_trait::async_trait;
use std::{collections::HashMap, sync::Arc};
use tokio::sync::RwLock;
use tracing::{event, Level};
use trust_dns_server::{
	authority::MessageResponseBuilder,
	client::{
		op::{Header, MessageType, OpCode, Query, ResponseCode},
		rr::{RData, Record, RecordType},
	},
	proto::xfer::DnsRequestOptions,
	resolver::{
		config::{NameServerConfig, Protocol, ResolverConfig, ResolverOpts},
		error::ResolveError,
		lookup::Lookup,
		TokioAsyncResolver,
	},
	server::{Request, RequestHandler, ResponseHandler, ResponseInfo},
};

use crate::{
	config::{Rule, SpecialUpstream, Upstream},
	dns_router::DnsRouter,
	proxy_runtime::{ProxyAsyncResolver, ProxyHandle},
};

#[derive(Debug, Clone)]
enum AsyncResolver {
	Directed(TokioAsyncResolver),
	Proxied(ProxyAsyncResolver),
}

#[derive(Debug)]
pub struct DnsHandler {
	router: DnsRouter,
	clients: Arc<RwLock<HashMap<Upstream, Result<AsyncResolver, ResolveError>>>>,
}

impl DnsHandler {
	pub fn new() -> Self {
		Self {
			router: DnsRouter::new(),
			clients: Arc::new(RwLock::new(HashMap::new())),
		}
	}

	pub fn add_rules(&mut self, rules: Vec<Rule>) {
		rules
			.into_iter()
			.rev()
			.enumerate()
			.for_each(|(priority, rule)| self.router.add_rule(rule, priority))
	}

	fn search_upstream(&self, request: &Request) -> Option<Arc<Upstream>> {
		let query = request.query();
		let record_type = query.query_type();
		let domain = query.name().to_string();
		self.router.search(domain, record_type)
	}

	async fn get_client(&self, upstream: Arc<Upstream>) -> Result<AsyncResolver, ResolveError> {
		let resolver = self.fast_get_client(upstream.clone()).await;
		if let Some(r) = resolver {
			r
		} else {
			self.slow_get_client(upstream).await
		}
	}
	async fn fast_get_client(
		&self,
		upstream: Arc<Upstream>,
	) -> Option<Result<AsyncResolver, ResolveError>> {
		let map = self.clients.clone();
		let map = map.read().await;
		let resolver = map.get(&upstream);
		resolver.map(|x| x.to_owned())
	}
	async fn slow_get_client(
		&self,
		upstream: Arc<Upstream>,
	) -> Result<AsyncResolver, ResolveError> {
		let map = self.clients.clone();
		let mut map = map.write().await;
		let resolver = map.entry((*upstream).clone()).or_insert_with(|| {
			let mut use_proxy = false;
			let name_server_config = match upstream.as_ref() {
				Upstream::UDP { udp } => NameServerConfig::new(udp.to_owned(), Protocol::Udp),
				Upstream::TCP { tcp } => NameServerConfig::new(tcp.to_owned(), Protocol::Tcp),
				Upstream::DoT { dot, domain } => {
					let mut c = NameServerConfig::new(dot.to_owned(), Protocol::Tls);
					c.tls_dns_name = Some(domain.to_owned());
					c
				}
				Upstream::DoH {
					doh,
					domain,
					socks5_proxy,
				} => {
					let mut c = NameServerConfig::new(doh.to_owned(), Protocol::Https);
					c.tls_dns_name = Some(domain.to_owned());
					if socks5_proxy.is_some() {
						use_proxy = true;
						c.bind_addr = socks5_proxy.to_owned();
					}
					c
				}
				_ => unreachable!(),
			};
			let mut resolver_config = ResolverConfig::new();
			resolver_config.add_name_server(name_server_config);
			let mut resolver_opts = ResolverOpts::default();
			resolver_opts.cache_size = 128;
			if use_proxy {
				ProxyAsyncResolver::new(resolver_config, resolver_opts, ProxyHandle)
					.map(AsyncResolver::Proxied)
			} else {
				TokioAsyncResolver::tokio(resolver_config, resolver_opts)
					.map(AsyncResolver::Directed)
			}
		});
		resolver.clone()
	}

	pub async fn query<R: ResponseHandler>(
		&self,
		request: &Request,
		mut response_handle: R,
	) -> Result<ResponseInfo, std::io::Error> {
		if let Some(upstream) = self.search_upstream(request) {
			let result = match upstream.as_ref() {
				Upstream::UDP { .. }
				| Upstream::TCP { .. }
				| Upstream::DoT { .. }
				| Upstream::DoH { .. } => {
					let resolver = self.get_client(upstream).await?;
					let query = request.query();
					let mut lookup_opt = DnsRequestOptions::default();
					lookup_opt.use_edns = request.edns().is_some();
					let result = match resolver {
						AsyncResolver::Directed(r) => {
							r.lookup(query.name(), query.query_type()).await?
						}
						AsyncResolver::Proxied(r) => {
							r.lookup(query.name(), query.query_type()).await?
						}
					};
					Some(result)
				}
				Upstream::IPv4 { ipv4 } => {
					if request.query().query_type() == RecordType::A {
						let query = Query::query(request.query().name().into(), RecordType::A);
						let result = Lookup::from_rdata(query, RData::A(*ipv4));
						Some(result)
					} else {
						None
					}
				}
				Upstream::IPv6 { ipv6 } => {
					if request.query().query_type() == RecordType::AAAA {
						let query = Query::query(request.query().name().into(), RecordType::AAAA);
						let result = Lookup::from_rdata(query, RData::AAAA(*ipv6));
						Some(result)
					} else {
						None
					}
				}
				Upstream::Special(SpecialUpstream::NODATA) => None,
				Upstream::Special(SpecialUpstream::NXDOMAIN) => {
					let response = MessageResponseBuilder::from_message_request(request);
					return response_handle
						.send_response(response.error_msg(request.header(), ResponseCode::NXDomain))
						.await;
				}
			};
			if let Some(result) = result {
				let answers = result.record_iter().collect::<Vec<&Record>>();
				let response = MessageResponseBuilder::from_message_request(request);
				let response_header = Header::response_from_request(request.header());
				let resp = response.build(response_header, answers, None, None, None);
				response_handle.send_response(resp).await
			} else {
				let response = MessageResponseBuilder::from_message_request(request);
				let response_header = Header::response_from_request(request.header());
				let resp = response.build_no_records(response_header);
				response_handle.send_response(resp).await
			}
		} else {
			event!(
				Level::WARN,
				"[{}] no upstream: {}",
				request.id(),
				request.query().name()
			);
			let response = MessageResponseBuilder::from_message_request(request);
			response_handle
				.send_response(response.error_msg(request.header(), ResponseCode::NXDomain))
				.await
		}
	}
}

#[async_trait]
impl RequestHandler for DnsHandler {
	async fn handle_request<R: ResponseHandler>(
		&self,
		request: &Request,
		mut response_handle: R,
	) -> ResponseInfo {
		let result = match request.message_type() {
			MessageType::Query => {
				event!(Level::DEBUG, "query received: {}", request.id());
				match request.op_code() {
					OpCode::Query => {
						event!(Level::DEBUG, "query: {:?}", request.id());
						self.query(request, response_handle).await
					}
					c => {
						event!(
							Level::WARN,
							"[{}] unimplemented op_code: {:?}",
							request.id(),
							c
						);
						let response = MessageResponseBuilder::from_message_request(request);
						response_handle
							.send_response(
								response.error_msg(request.header(), ResponseCode::NotImp),
							)
							.await
					}
				}
			}
			MessageType::Response => {
				event!(
					Level::WARN,
					"got a response as a request from id: {}",
					request.id()
				);
				let response = MessageResponseBuilder::from_message_request(request);
				response_handle
					.send_response(response.error_msg(request.header(), ResponseCode::FormErr))
					.await
			}
		};

		return match result {
			Err(e) => {
				event!(Level::ERROR, "request failed: {}", e);
				// copy from ResponseInfo::serve_failed()
				let mut header = Header::new();
				header.set_response_code(ResponseCode::ServFail);
				header.into()
			}
			Ok(info) => info,
		};
	}
}
