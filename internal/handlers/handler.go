package handlers

import (
	"atlas/internal/address"
	"atlas/internal/assets"
	"atlas/internal/caching"
	"atlas/pkg/goccm"
	"atlas/pkg/logging"
	"bufio"
	"encoding/base64"
	"encoding/json"
	"github.com/tidwall/gjson"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type handlerExtraArguments struct {
	proxyAuth gjson.Result
	arguments []byte
	c         goccm.ConcurrencyManager
	start     time.Time
}

// Handle handles the connection
func Handle(c net.Conn) {
	var proxyAuth gjson.Result
	arguments := make(map[string]string)

	start := time.Now()

	// Close the connection when the function returns
	defer func() {
		err := c.Close()

		if err != nil {
			logging.Logger.Error().
				Err(err).
				Msg("Error closing connection")
		}
	}()

	// Set a deadline for reading. Read operation will fail if no data
	request, err := http.ReadRequest(bufio.NewReader(c))
	if err != nil {
		logging.Logger.Error().
			Err(err).
			Msg("Error reading request")

		response := http.Response{
			StatusCode: http.StatusBadRequest,
			Status:     http.StatusText(http.StatusBadRequest),
		}

		response.Proto = "HTTP/1.1"
		response.ProtoMajor = 1
		response.ProtoMinor = 1

		response.Header = make(http.Header)
		response.Header.Set("Connection", "close")
		response.Header.Set("Content-Length", "0")

		response.ContentLength = 0
		response.Close = true

		if err := response.Write(c); err != nil {
			logging.Logger.Error().
				Err(err).
				Msg("Error writing response")
		}

		return
	}

	proxyAuthorization := request.Header.Get("Proxy-Authorization")
	if proxyAuthorization == "" {
		response := http.Response{
			StatusCode: http.StatusUnauthorized,
			Status:     http.StatusText(http.StatusUnauthorized),
		}

		response.Proto = "HTTP/1.1"
		response.ProtoMajor = 1
		response.ProtoMinor = 1

		response.Header = make(http.Header)
		response.Header.Set("Connection", "close")
		response.Header.Set("Content-Length", "0")

		response.ContentLength = 0
		response.Close = true

		if err := response.Write(c); err != nil {
			logging.Logger.Error().
				Err(err).
				Msg("Error writing response")
		}
		return
	}

	// Decode the proxy authorization header and split it into username and password parts (separated by a colon)
	decodedProxyAuthorization, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(proxyAuthorization, "Basic ", ""))
	if err != nil {
		response := http.Response{
			StatusCode: http.StatusUnauthorized,
			Status:     http.StatusText(http.StatusUnauthorized),
		}

		response.Proto = "HTTP/1.1"
		response.ProtoMajor = 1
		response.ProtoMinor = 1

		response.Header = make(http.Header)
		response.Header.Set("Connection", "close")
		response.Header.Set("Content-Length", "0")

		response.ContentLength = 0
		response.Close = true

		if err := response.Write(c); err != nil {
			logging.Logger.Error().
				Err(err).
				Msg("Error writing response")
		}
		return
	}

	username := strings.Split(string(decodedProxyAuthorization), ":")[0]
	password := strings.Split(string(decodedProxyAuthorization), ":")[1]

	// Check if the password contains arguments
	pattern := `_([^_]+)-([^_@]+)`
	regex := regexp.MustCompile(pattern)
	matches := regex.FindAllStringSubmatch(password, -1)

	for _, match := range matches {
		arguments[match[1]] = match[2]
	}

	// Remove the arguments from the password
	if len(arguments) != 0 {
		for key, value := range arguments {
			password = strings.ReplaceAll(password, "_"+key+"-"+value, "")
		}
	}

	// Check if the account exists
	details, ok := caching.Accounts.Get(username)
	if !ok {
		response := http.Response{
			StatusCode: http.StatusUnauthorized,
			Status:     http.StatusText(http.StatusUnauthorized),
		}

		response.Proto = "HTTP/1.1"
		response.ProtoMajor = 1
		response.ProtoMinor = 1

		response.Header = make(http.Header)
		response.Header.Set("Connection", "close")
		response.Header.Set("Content-Length", "0")

		response.ContentLength = 0
		response.Close = true

		if err := response.Write(c); err != nil {
			logging.Logger.Error().
				Err(err).
				Msg("Error writing response")
		}

		return
	} else {
		// Check if the password is correct
		if password != details.(gjson.Result).Get("password").String() {
			response := http.Response{
				StatusCode: http.StatusUnauthorized,
				Status:     http.StatusText(http.StatusUnauthorized),
			}

			response.Proto = "HTTP/1.1"
			response.ProtoMajor = 1
			response.ProtoMinor = 1

			response.Header = make(http.Header)
			response.Header.Set("Connection", "close")
			response.Header.Set("Content-Length", "0")

			response.ContentLength = 0
			response.Close = true

			if err := response.Write(c); err != nil {
				logging.Logger.Error().
					Err(err).
					Msg("Error writing response")
			}

			return
		} else {
			proxyAuth = details.(gjson.Result)
		}
	}

	// Check if the account is expired
	if time.Unix(proxyAuth.Get("expiry").Int(), 0).Before(time.Now()) {
		response := http.Response{
			StatusCode: http.StatusPaymentRequired,
			Status:     http.StatusText(http.StatusPaymentRequired),
		}

		response.Proto = "HTTP/1.1"
		response.ProtoMajor = 1
		response.ProtoMinor = 1

		response.Header = make(http.Header)
		response.Header.Set("Connection", "close")
		response.Header.Set("Content-Length", "0")

		response.ContentLength = 0
		response.Close = true

		if err := response.Write(c); err != nil {
			logging.Logger.Error().
				Err(err).
				Msg("Error writing response")
		}

		return
	}

	request.Header.Del("Proxy-Connection")
	request.Header.Del("Proxy-Authenticate")
	request.Header.Del("Proxy-Authorization")

	if request.Header.Get("Connection") == "close" {
		request.Close = false
	}

	request.Header.Del("Connection")
	request.Header.Del("Accept-Encoding")

	var cidr *net.IPNet
	var ipv6Address net.IP

	country, ok := arguments["country"]
	if ok {
		if subnet, ok := assets.Subnets[strings.ToLower(country)]; ok {
			_, cidr, err = net.ParseCIDR(subnet)
			if err != nil {
				logging.Logger.Error().
					Err(err).
					Msg("Error parsing CIDR")

				response := http.Response{
					StatusCode: http.StatusInternalServerError,
					Status:     http.StatusText(http.StatusInternalServerError),
				}

				response.Proto = "HTTP/1.1"
				response.ProtoMajor = 1
				response.ProtoMinor = 1

				response.Header = make(http.Header)
				response.Header.Set("Connection", "close")
				response.Header.Set("Content-Length", "0")

				response.ContentLength = 0
				response.Close = true

				if err := response.Write(c); err != nil {
					logging.Logger.Error().
						Err(err).
						Msg("Error writing response")
				}

				return
			}
		}
	}
	if cidr == nil {
		_, cidr, err = net.ParseCIDR(assets.Subnets["default"])
		if err != nil {
			logging.Logger.Error().
				Err(err).
				Msg("Error parsing CIDR")

			response := http.Response{
				StatusCode: http.StatusInternalServerError,
				Status:     http.StatusText(http.StatusInternalServerError),
			}

			response.Proto = "HTTP/1.1"
			response.ProtoMajor = 1
			response.ProtoMinor = 1

			response.Header = make(http.Header)
			response.Header.Set("Connection", "close")
			response.Header.Set("Content-Length", "0")

			response.ContentLength = 0
			response.Close = true

			if err := response.Write(c); err != nil {
				logging.Logger.Error().
					Err(err).
					Msg("Error writing response")
			}

			return
		}
	}

	session, ok := arguments["session"]
	if ok {
		ipv6Address = address.RandomSeededIPv6(cidr, session)
	} else {
		ipv6Address = address.RandomIPv6(cidr)
	}

	argumentsStr, _ := json.Marshal(arguments)

	caching.Concurrent[proxyAuth.Get("username").String()].Wait()
	logging.Logger.Debug().
		RawJSON("arguments", argumentsStr).
		Int32("threads", caching.Concurrent[proxyAuth.Get("username").String()].RunningCount()).
		Str("ipv6", ipv6Address.String()).
		Str("host", request.Host).
		Str("user", proxyAuth.Get("username").String()).
		Str("took", time.Since(start).String()).
		Msg("Forwarding request to subnet.")

	extraArgs := handlerExtraArguments{
		proxyAuth: proxyAuth,
		arguments: argumentsStr,
		c:         caching.Concurrent[proxyAuth.Get("username").String()],
		start:     start,
	}

	if request.Method == "CONNECT" {
		// Handle HTTPS requests
		err := handleHTTPS(c, request, ipv6Address, extraArgs)
		if err != nil {
			logging.Logger.Error().
				Err(err).
				Msg("Error handling request")

			response := http.Response{
				StatusCode: http.StatusInternalServerError,
				Status:     http.StatusText(http.StatusInternalServerError),
			}

			response.Proto = "HTTP/1.1"
			response.ProtoMajor = 1
			response.ProtoMinor = 1

			response.Header = make(http.Header)
			response.Header.Set("Connection", "close")
			response.Header.Set("Content-Length", "0")

			response.ContentLength = 0
			response.Close = true

			if err := response.Write(c); err != nil {
				logging.Logger.Error().
					Err(err).
					Msg("Error writing response")
			}

			return
		}
	} else {
		// Handle HTTP requests
		err := handleHTTP(c, request, ipv6Address, extraArgs)
		if err != nil {
			logging.Logger.Error().
				Err(err).
				Msg("Error handling request")

			response := http.Response{
				StatusCode: http.StatusInternalServerError,
				Status:     http.StatusText(http.StatusInternalServerError),
			}

			response.Proto = "HTTP/1.1"
			response.ProtoMajor = 1
			response.ProtoMinor = 1

			response.Header = make(http.Header)
			response.Header.Set("Connection", "close")
			response.Header.Set("Content-Length", "0")

			response.ContentLength = 0
			response.Close = true

			if err := response.Write(c); err != nil {
				logging.Logger.Error().
					Err(err).
					Msg("Error writing response")
			}

			return
		}
	}
}
