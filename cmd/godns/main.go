package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"

	"github.com/dhcmrlchtdj/godns/internal/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack // nolint:reassign
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	dnsServer := server.NewDnsServer(ctx)
	dnsServer.ParseArgs()

	level, err := zerolog.ParseLevel(dnsServer.Config.LogLevel)
	if err != nil {
		panic(err)
	}
	zerolog.SetGlobalLevel(level)

	dnsServer.SetupRouter()
	dnsServer.SetupServer()
	dnsServer.SetupPprof()
	dnsServer.Start()
}
