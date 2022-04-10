# lego-dnsserver 
хук для https://go-acme.github.io/lego/dns/httpreq/ (чтобы выписать wildcard letsencrypt certificate)

поднимает локальный днс сервер, удобно например в связке с coredns:
```Corefile
_acme-challenge.foo.com {
    forward . 127.0.0.1:5353
}
```

```text
Usage of ./dist/darwin/amd64/lego-dnsserver:
  -listen-dns string
        Listen addr for serve dns records (default "127.0.0.1:5352")
  -listen-http string
        Listen addr for serve lego httpreq httpreq https://go-acme.github.io/lego/dns/httpreq/ (default "127.0.0.1:18888")
```
