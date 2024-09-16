package handlers

import (
	"atlas/internal/caching"
	"atlas/pkg/freebind"
	"atlas/pkg/logging"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

func handleHTTPS(c net.Conn, r *http.Request, ip net.IP, args handlerExtraArguments) error {
	dialer := net.Dialer{
		LocalAddr: &net.TCPAddr{
			IP: ip,
		},
		Control: freebind.ControlFreeBind,
	}

	conn, err := dialer.Dial("tcp", r.Host)
	if err != nil {
		args.c.Done()
		return err
	}

	defer conn.Close()

	response := &http.Response{
		Status:        "200 OK",
		StatusCode:    200,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          ioutil.NopCloser(strings.NewReader("")),
		ContentLength: 0,
	}

	if err := response.Write(c); err != nil {
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

	go io.Copy(conn, c)
	if _, err := io.Copy(c, conn); err != nil {
		return err
	}

	return nil
}
