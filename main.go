package main

import (
	"fmt"

	"github.com/miekg/dns"

	"github.com/dhcmrlchtdj/shunt/client"
	"github.com/dhcmrlchtdj/shunt/config"
)

type Shunt struct {
	config config.Config
	server dns.Server
	client client.DNSClient
}

func main() {
	dnsMux := dns.NewServeMux()
	s := Shunt{
		server: dns.Server{
			Addr:    ":1053",
			Net:     "udp",
			Handler: dnsMux,
		},
	}
	dnsMux.HandleFunc(".", s.handleRequest)
	s.config.Load("./config.json")
	s.client.Init(s.config.Forward)

	fmt.Println("Starting at 1053")
	err := s.server.ListenAndServe()
	if err != nil {
		panic(err)
	}
	defer s.server.Shutdown()
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
