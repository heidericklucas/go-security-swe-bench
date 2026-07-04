package fetch

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/netip"
	"time"
)

// Resolver maps a host to candidate IPs for the dial decision.
type Resolver func(ctx context.Context, host string) ([]netip.Addr, error)

// DialFunc opens a connection to a specific resolved IP:port.
type DialFunc func(ctx context.Context, network string, ip netip.Addr, port string) (net.Conn, error)

func defaultResolver(ctx context.Context, host string) ([]netip.Addr, error) {
	return net.DefaultResolver.LookupNetIP(ctx, "ip", host)
}

func defaultDial(ctx context.Context, network string, ip netip.Addr, port string) (net.Conn, error) {
	var d net.Dialer
	return d.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
}

// Fetcher performs a server-side GET, following redirects.
type Fetcher struct {
	Resolver Resolver
	Dial     DialFunc
	Timeout  time.Duration
}

func New() *Fetcher {
	return &Fetcher{Resolver: defaultResolver, Dial: defaultDial, Timeout: 5 * time.Second}
}

func (f *Fetcher) Get(ctx context.Context, rawURL string) ([]byte, error) {
	client := &http.Client{
		Timeout:   f.Timeout,
		Transport: &http.Transport{DialContext: f.dialContext},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(io.LimitReader(resp.Body, 1<<20))
}

// dialContext resolves addr and dials the first resolved address.
func (f *Fetcher) dialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	ips, err := f.Resolver(ctx, host)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, errors.New("no such host")
	}
	return f.Dial(ctx, network, ips[0], port)
}
