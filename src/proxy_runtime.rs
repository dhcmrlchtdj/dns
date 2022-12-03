use std::{
	future::Future,
	net::SocketAddr,
	pin::Pin,
	task::{Context, Poll},
};
use tokio::{
	io::{AsyncRead, AsyncWrite, ReadBuf},
	net::TcpStream,
};
use tokio_socks::tcp::Socks5Stream;
use trust_dns_server::{
	proto::{
		error::ProtoError,
		tcp::{Connect, DnsTcpStream},
		TokioTime,
	},
	resolver::{
		name_server::{GenericConnection, GenericConnectionProvider, RuntimeProvider, Spawn},
		AsyncResolver,
	},
};

pub type ProxyConnectionProvider = GenericConnectionProvider<ProxyRuntime>;
pub type ProxyAsyncResolver = AsyncResolver<GenericConnection, ProxyConnectionProvider>;

// AsyncResolver::new(resolver_config, resolver_opts, ProxyHandle)
// ProxyAsyncResolver::new(resolver_config, resolver_opts, ProxyHandle)

#[derive(Clone)]
pub struct ProxyHandle;
impl Spawn for ProxyHandle {
	fn spawn_bg<F>(&mut self, future: F)
	where
		F: Future<Output = Result<(), ProtoError>> + Send + 'static,
	{
		let _join = tokio::spawn(future);
	}
}

#[derive(Clone)]
pub struct ProxyRuntime;
impl RuntimeProvider for ProxyRuntime {
	type Handle = ProxyHandle;
	type Tcp = ProxyTcpStream;
	type Timer = TokioTime;
	type Udp = tokio::net::UdpSocket;
}

pub struct ProxyTcpStream {
	inner: Socks5Stream<TcpStream>,
}

impl futures_io::AsyncRead for ProxyTcpStream {
	fn poll_read(
		self: Pin<&mut Self>,
		cx: &mut Context<'_>,
		buf: &mut [u8],
	) -> Poll<std::io::Result<usize>> {
		let mut buf = ReadBuf::new(buf);
		let polled = AsyncRead::poll_read(Pin::new(&mut self.get_mut().inner), cx, &mut buf);
		polled.map_ok(|_| buf.filled().len())
	}
}
impl futures_io::AsyncWrite for ProxyTcpStream {
	fn poll_write(
		self: Pin<&mut Self>,
		cx: &mut Context<'_>,
		buf: &[u8],
	) -> Poll<std::io::Result<usize>> {
		AsyncWrite::poll_write(Pin::new(&mut self.get_mut().inner), cx, buf)
	}
	fn poll_flush(self: Pin<&mut Self>, cx: &mut Context<'_>) -> Poll<std::io::Result<()>> {
		AsyncWrite::poll_flush(Pin::new(&mut self.get_mut().inner), cx)
	}
	fn poll_close(self: Pin<&mut Self>, cx: &mut Context<'_>) -> Poll<std::io::Result<()>> {
		AsyncWrite::poll_shutdown(Pin::new(&mut self.get_mut().inner), cx)
	}
}

#[async_trait::async_trait]
impl Connect for ProxyTcpStream {
	async fn connect_with_bind(
		addr: SocketAddr,
		bind_addr: Option<SocketAddr>,
	) -> std::io::Result<Self> {
		if let Some(socks5_proxy) = bind_addr {
			match Socks5Stream::connect(socks5_proxy, addr).await {
				Ok(inner) => Ok(Self { inner }),
				Err(err) => Err(futures_io::Error::new(futures_io::ErrorKind::Other, err)),
			}
		} else {
			unreachable!()
		}
	}
}
impl DnsTcpStream for ProxyTcpStream {
	type Time = TokioTime;
}
