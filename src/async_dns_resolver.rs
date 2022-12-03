use std::fmt::Debug;

use trust_dns_server::{
	client::rr::LowerName,
	proto::rr::RecordType,
	resolver::{
		error::ResolveError,
		lookup::Lookup,
		name_server::{GenericConnection, GenericConnectionProvider, TokioRuntime},
		AsyncResolver,
	},
};

use crate::proxy_runtime::ProxyRuntime;

#[async_trait::async_trait]
pub trait MyAsyncResolver: Send + Sync + Debug {
	async fn resolve(
		&self,
		name: &LowerName,
		record_type: RecordType,
	) -> Result<Lookup, ResolveError>;
}

#[async_trait::async_trait]
impl MyAsyncResolver for AsyncResolver<GenericConnection, GenericConnectionProvider<ProxyRuntime>> {
	async fn resolve(
		&self,
		name: &LowerName,
		record_type: RecordType,
	) -> Result<Lookup, ResolveError> {
		self.lookup(name, record_type).await
	}
}

#[async_trait::async_trait]
impl MyAsyncResolver for AsyncResolver<GenericConnection, GenericConnectionProvider<TokioRuntime>> {
	async fn resolve(
		&self,
		name: &LowerName,
		record_type: RecordType,
	) -> Result<Lookup, ResolveError> {
		self.lookup(name, record_type).await
	}
}
