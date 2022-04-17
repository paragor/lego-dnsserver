package pkg

import (
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"testing"
)

var listenAddr = "127.0.0.1:5352"

func TestDnsProvider(t *testing.T) {
	provider, err := NewDNSServer(listenAddr)
	assert.Nil(t, err)

	fqdn := "_acme-challenge.domain."
	txtExpectedValue := "LHDhK3oGRvkiefQnx7OOczTY5Tic_xZ6HcMOc_gmtoM"

	err = provider.Present(fqdn, txtExpectedValue)
	assert.Nil(t, err)

	r, err := makeDnsQuery(fqdn, dns.TypeTXT, listenAddr)
	assert.Nil(t, err)
	assert.Equal(t, txtExpectedValue, r.Answer[0].(*dns.TXT).Txt[0])

	r, err = makeDnsQuery("bar.com.", dns.TypeTXT, listenAddr)
	assert.Nil(t, err)
	assert.Equal(t, dns.RcodeNameError, r.MsgHdr.Rcode)

	err = provider.CleanUp()
	assert.Nil(t, err)
}

func makeDnsQuery(fqdn string, dnsType uint16, dnsServer string) (*dns.Msg, error) {
	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(fqdn, dnsType)
	r, _, err := c.Exchange(m, dnsServer)
	return r, err
}
