package pkg

import (
	"errors"
	"github.com/miekg/dns"
	"log"
	"net"
	"net/netip"
)

type DNSServer struct {
	addr   string
	server *dns.Server
}

func NewDNSServer(listenAddr string) (*DNSServer, error) {
	_, err := netip.ParseAddrPort(listenAddr)
	if err != nil {
		return nil, err
	}
	return &DNSServer{addr: listenAddr}, nil
}

func (s *DNSServer) Present(fqdn, value string) error {
	addr, err := netip.ParseAddrPort(s.addr)
	if err != nil {
		return err
	}

	if s.IsUp() {
		err = s.CleanUp()
		if err != nil {
			return err
		}
	}

	conn, err := net.ListenUDP("udp", net.UDPAddrFromAddrPort(addr))
	if err != nil {
		return err
	}
	handler := &dnsHandler{acmeResponse: value, fqdn: fqdn}
	s.server = &dns.Server{Handler: handler, PacketConn: conn}

	startCh := make(chan struct{}, 1)
	errCh := make(chan error, 1)
	s.server.NotifyStartedFunc = func() {
		startCh <- struct{}{}
	}
	go func() {
		err := s.server.ActivateAndServe()
		if err != nil {
			errCh <- err
			return
		}
	}()
	select {
	case err := <-errCh:
		return err
	case <-startCh:
	}

	return nil
}
func (s *DNSServer) IsUp() bool {
	return s.server != nil
}

func (s *DNSServer) CleanUp() error {
	if s.server == nil {
		return errors.New("dns server is not running")
	}
	err := s.server.Shutdown()
	s.server = nil
	return err
}

type dnsHandler struct {
	acmeResponse string
	fqdn         string
}

func (h *dnsHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	if len(r.Question) == 0 {
		log.Println("invalid dns request - no questions. force close connection")
		w.Close()
		return
	}
	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true
	m.RecursionAvailable = false
	question := r.Question[0]
	if question.Name == h.fqdn && question.Qtype == dns.TypeTXT {
		m.Answer = append(m.Answer, &dns.TXT{
			Hdr: dns.RR_Header{
				Name:   question.Name,
				Rrtype: dns.TypeTXT,
				Class:  dns.ClassINET,
			},
			Txt: []string{h.acmeResponse},
		})
		log.Printf("localserver dns response: found, request: %#v\n", question)
	} else {
		m.Rcode = dns.RcodeNameError
		log.Printf("localserver dns response: not found, request: %#v\n", question)
	}
	if err := w.WriteMsg(m); err != nil {
		log.Println(err.Error())
	}
}
