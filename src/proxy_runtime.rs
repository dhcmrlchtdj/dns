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

pub type ProxyConnection = GenericConnection;
pub type ProxyConnectionProvider<P: MaybeSocketAddr> = GenericConnectionProvider<ProxyRuntime<P>>;
pub type ProxyAsyncResolver<P: MaybeSocketAddr> =
    AsyncResolver<ProxyConnection, ProxyConnectionProvider<P>>;

// AsyncResolver::new(resolver_config, resolver_opts, ProxyHandle)
// ProxyAsyncResolver::new(resolver_config, resolver_opts, ProxyHandle)

pub trait MaybeSocketAddr {
    fn get_proxy_addr() -> Option<SocketAddr>;
}

pub struct OptAddr(Option((String, u16)));
impl MaybeSocketAddr for OptAddr {
    fn get_proxy_addr() -> Option<SocketAddr> {
        
    }
}

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
pub struct ProxyRuntime<P: MaybeSocketAddr>(P);
impl<P: MaybeSocketAddr> RuntimeProvider for ProxyRuntime<P> {
    type Handle = ProxyHandle;
    type Tcp = ProxyTcpStream<P>;
    type Timer = TokioTime;
    type Udp = tokio::net::UdpSocket;
}

pub struct ProxyTcpStream<P: MaybeSocketAddr> {
    inner: Socks5Stream<TcpStream>,
    proxy: P,
}

impl<P: MaybeSocketAddr> futures_io::AsyncRead for ProxyTcpStream<P> {
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
impl<P: MaybeSocketAddr> futures_io::AsyncWrite for ProxyTcpStream<P> {
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
impl<P: MaybeSocketAddr> Connect for ProxyTcpStream<P> {
    async fn connect_with_bind(
        addr: SocketAddr,
        _bind_addr: Option<SocketAddr>,
    ) -> std::io::Result<Self> {
        let proxy_addr = ("127.0.0.1", 1080);
        match Socks5Stream::connect(proxy_addr, addr).await {
            Ok(inner) => Ok(Self { inner, proxy: P }),
            Err(err) => Err(futures_io::Error::new(futures_io::ErrorKind::Other, err)),
        }
    }
}
impl<P: MaybeSocketAddr> DnsTcpStream for ProxyTcpStream<P> {
    type Time = TokioTime;
}
