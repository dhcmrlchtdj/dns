package main

import (
	"fmt"

	"github.com/miekg/dns"

	"github.com/dhcmrlchtdj/shunt/client"
)

type Shunt struct {
	server dns.Server
	client client.DNSClient
}

func main() {
	s := Shunt{
		server: dns.Server{Addr: ":1053", Net: "udp"},
	}
	s.client.LoadConfig()

	fmt.Println("Starting at 1053")
	err := s.server.ListenAndServe()
	if err != nil {
		panic(err)
	}
	defer s.server.Shutdown()

	dns.HandleFunc(".", s.handleRequest)
}

func (s *Shunt) handleRequest(w dns.ResponseWriter, query *dns.Msg) {
	println("handle request")
	m := new(dns.Msg)
	m.SetReply(query)

	if query.Opcode == dns.OpcodeQuery {
		s.Query(m)
	}

	w.WriteMsg(m)
}

func (s *Shunt) Query(m *dns.Msg) {
	for _, q := range m.Question {
		answers := s.client.Query(q.Name, q.Qtype)
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
