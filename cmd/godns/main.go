package main

import (
	"strconv"

	"github.com/morikuni/failure"
	"github.com/rs/zerolog"

	"github.com/dhcmrlchtdj/godns/internal/server"
)

func main() {
	zerolog.ErrorStackMarshaler = marshalStack // nolint:reassign
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	dnsServer := server.NewDnsServer()
	dnsServer.SetupContext()
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

func marshalStack(err error) interface{} {
	cs, ok := failure.CallStackOf(err)
	if !ok {
		return nil
	}
	frames := cs.Frames()
	out := make([]map[string]string, 0, len(frames))
	for _, frame := range frames {
		out = append(out, map[string]string{
			"path": frame.Path(),
			"line": strconv.Itoa(frame.Line()),
			"func": frame.Func(),
		})
	}
	return out
}
