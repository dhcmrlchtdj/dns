use crate::{
    config::{Rule, SpecialUpstream, Upstream},
    dns_router::DnsRouter,
    proxy_runtime::{ProxyAsyncResolver, ProxyHandle},
};
use async_trait::async_trait;
use std::{collections::HashMap, sync::Arc};
use tokio::sync::Mutex;
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
    },
    server::{Request, RequestHandler, ResponseHandler, ResponseInfo},
};

#[derive(Debug)]
pub struct DnsHandler {
    router: DnsRouter,
    clients: Arc<Mutex<HashMap<Upstream, Result<ProxyAsyncResolver, ResolveError>>>>,
}

impl DnsHandler {
    pub fn new() -> Self {
        Self {
            router: DnsRouter::new(),
            clients: Arc::new(Mutex::new(HashMap::new())),
        }
    }

    pub fn add_rules(&mut self, rules: Vec<Rule>) {
        rules
            .into_iter()
            .enumerate()
            .for_each(|(index, rule)| self.router.add_rule(rule, index))
    }

    fn search_upstream(&self, request: &Request) -> Option<Upstream> {
        let query = request.query();
        let record_type = query.query_type();
        let domain = query.name().to_string();
        self.router.search(domain, record_type)
    }

    async fn get_client(&self, upstream: Upstream) -> Result<ProxyAsyncResolver, ResolveError> {
        let map = self.clients.clone();
        let mut map = map.lock().await;
        let resolver = map.entry(upstream.clone()).or_insert_with(|| {
            let name_server_config = match upstream {
                Upstream::UDP { udp } => NameServerConfig {
                    socket_addr: udp,
                    protocol: Protocol::Udp,
                    tls_dns_name: None,
                    trust_nx_responses: true,
                    tls_config: None,
                    bind_addr: None,
                },
                Upstream::TCP { tcp } => NameServerConfig {
                    socket_addr: tcp,
                    protocol: Protocol::Tcp,
                    tls_dns_name: None,
                    trust_nx_responses: true,
                    tls_config: None,
                    bind_addr: None,
                },
                Upstream::DoT { dot, domain } => NameServerConfig {
                    socket_addr: dot,
                    protocol: Protocol::Tls,
                    tls_dns_name: Some(domain),
                    trust_nx_responses: true,
                    tls_config: None,
                    bind_addr: None,
                },
                Upstream::DoH { doh, domain } => NameServerConfig {
                    socket_addr: doh,
                    protocol: Protocol::Https,
                    tls_dns_name: Some(domain),
                    trust_nx_responses: true,
                    tls_config: None,
                    bind_addr: None,
                },
                _ => unreachable!(),
            };
            let mut resolver_config = ResolverConfig::new();
            resolver_config.add_name_server(name_server_config);
            let mut resolver_opts = ResolverOpts::default();
            resolver_opts.cache_size = 512;
            ProxyAsyncResolver::new(resolver_config, resolver_opts, ProxyHandle)
        });
        resolver.to_owned()
    }

    pub async fn query<R: ResponseHandler>(
        &self,
        request: &Request,
        mut response_handle: R,
    ) -> Result<ResponseInfo, std::io::Error> {
        if let Some(upstream) = self.search_upstream(request) {
            let result = match upstream {
                Upstream::UDP { .. } => {
                    let resolver = self.get_client(upstream).await?;
                    let query = request.query();
                    let mut lookup_opt = DnsRequestOptions::default();
                    lookup_opt.use_edns = request.edns().is_some();
                    let result = resolver
                        .lookup(query.name(), query.query_type(), lookup_opt)
                        .await?;
                    Some(result)
                }
                Upstream::TCP { .. } => {
                    let resolver = self.get_client(upstream).await?;
                    let query = request.query();
                    let mut lookup_opt = DnsRequestOptions::default();
                    lookup_opt.use_edns = request.edns().is_some();
                    let result = resolver
                        .lookup(query.name(), query.query_type(), lookup_opt)
                        .await?;
                    Some(result)
                }
                Upstream::DoT { .. } => {
                    let resolver = self.get_client(upstream).await?;
                    let query = request.query();
                    let mut lookup_opt = DnsRequestOptions::default();
                    lookup_opt.use_edns = request.edns().is_some();
                    let result = resolver
                        .lookup(query.name(), query.query_type(), lookup_opt)
                        .await?;
                    Some(result)
                }
                Upstream::DoH { .. } => {
                    let resolver = self.get_client(upstream).await?;
                    let query = request.query();
                    let mut lookup_opt = DnsRequestOptions::default();
                    lookup_opt.use_edns = request.edns().is_some();
                    let result = resolver
                        .lookup(query.name(), query.query_type(), lookup_opt)
                        .await?;
                    Some(result)
                }
                Upstream::IPv4 { ipv4 } => {
                    if request.query().query_type() == RecordType::A {
                        let query = Query::query(request.query().name().into(), RecordType::A);
                        let result = Lookup::from_rdata(query, RData::A(ipv4));
                        Some(result)
                    } else {
                        None
                    }
                }
                Upstream::IPv6 { ipv6 } => {
                    if request.query().query_type() == RecordType::AAAA {
                        let query = Query::query(request.query().name().into(), RecordType::AAAA);
                        let result = Lookup::from_rdata(query, RData::AAAA(ipv6));
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