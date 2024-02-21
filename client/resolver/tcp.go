package resolver

import (
	"errors"
	"strings"

	"google.golang.org/grpc/resolver"
)

const scheme = "tcp"

type TcpBuilder struct{}

func (TcpBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	if target.Endpoint() == "" && opts.Dialer == nil {
		return nil, errors.New("tcp: received empty target in Build()")
	}
	addresses := []resolver.Address{}
	parsedTarget := ParseTarget(target.Endpoint())
	for _, t := range strings.Split(parsedTarget.URL.Host, ",") {
		addresses = append(addresses, resolver.Address{Addr: t})
	}
	r := &tcpResolver{
		target:    target,
		cc:        cc,
		addresses: addresses,
	}
	r.start()
	return r, nil
}

func (TcpBuilder) Scheme() string {
	return scheme
}

type tcpResolver struct {
	target    resolver.Target
	cc        resolver.ClientConn
	addresses []resolver.Address
}

func (r tcpResolver) start() {
	r.cc.UpdateState(resolver.State{Addresses: r.addresses})
}

func (tcpResolver) ResolveNow(o resolver.ResolveNowOptions) {}

func (tcpResolver) Close() {}

func init() {
	resolver.Register(TcpBuilder{})
}
