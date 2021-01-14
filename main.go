package main

import (
	"fmt"

	"github.com/miekg/dns"

	"github.com/dhcmrlchtdj/shunt/client"
)

var dnsClient = new(client.DNSClient)

func main() {
	dnsClient.LoadConfig()

	dns.HandleFunc(".", handleRequest)

	server := &dns.Server{Addr: ":1053", Net: "udp"}
	fmt.Println("Starting at 1053")
	err := server.ListenAndServe()
	defer server.Shutdown()
	if err != nil {
		panic(err)
	}
}

func handleRequest(w dns.ResponseWriter, query *dns.Msg) {
	println("handle request")
	m := new(dns.Msg)
	m.SetReply(query)

	if query.Opcode == dns.OpcodeQuery {
		Query(m)
	}

	w.WriteMsg(m)
}

func Query(m *dns.Msg) {
	for _, q := range m.Question {
		answers := dnsClient.Query(q.Name, q.Qtype)
		for _, ans := range answers {
			record := fmt.Sprintf("%s %d %s %s", ans.Name, ans.TTL, dns.Type(ans.Type).String(), ans.Data)
			rr, err := dns.NewRR(record)
			if err != nil {
				panic(err)
			}
			m.Answer = append(m.Answer, rr)
		}
	}
}
