package handlers

import (
	"atlas/internal/caching"
	"atlas/pkg/freebind"
	"atlas/pkg/logging"
	"context"
	"net"
	"net/http"
	"time"
)

func handleHTTP(c net.Conn, r *http.Request, ip net.IP, args handlerExtraArguments) error {
	r.RequestURI = ""

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network string, addr string) (net.Conn, error) {
				d := net.Dialer{
					LocalAddr: &net.TCPAddr{
						IP: ip,
					},
					Control: freebind.ControlFreeBind,
				}

				return d.Dial(network, addr)
			},
		},
	}

	resp, err := client.Do(r)
	if err != nil {
		args.c.Done()
		return err
	}

	args.c.Done()
	logging.Logger.Debug().
		RawJSON("arguments", args.arguments).
		Int32("threads", caching.Concurrent[args.proxyAuth.Get("username").String()].RunningCount()).
		Str("ipv6", ip.String()).
		Str("host", r.Host).
		Str("user", args.proxyAuth.Get("username").String()).
		Str("took", time.Since(args.start).String()).
		Msg("Completed request.")

	defer resp.Body.Close()
	if err := resp.Write(c); err != nil {
		return err
	}

	return nil
}
