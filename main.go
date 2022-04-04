package main

import (
	"github.com/dhcmrlchtdj/godns/server"
)

func main() {
	dnsServer := server.NewDnsServer()
	dnsServer.ParseArgs()
	dnsServer.InitRouter()
	dnsServer.InitServer()
	dnsServer.Start()
}
