//  Package server implements the http frontend
package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dimfeld/httptreemux"

	"github.com/go-spatial/tegola"
	"github.com/go-spatial/tegola/atlas"
	"github.com/go-spatial/tegola/internal/log"
	"github.com/go-spatial/tegola/dict"
	"crypto/tls"
)

const (
	// MaxTileSize is 500k. Currently just throws a warning when tile
	// is larger than MaxTileSize
	MaxTileSize = 500000
)

const (
	configKeyPort = "port"
	configKeySSLKey = "ssl_key"
	configKeySSLCert = "ssl_cert"
)

var (
	// Version is the version of the software, this should be set by the main program, before starting up.
	// It is used by various Middleware to determine the version.
	Version string = "Version Not Set"

	// HostName is the name of the host to use for construction of URLS.
	// configurable via the tegola config.toml file (set in main.go)
	HostName string

	// Port is the port the server is listening on, used for construction of URLS.
	// configurable via the tegola config.toml file (set in main.go)
	Port string

	// SSLCert is a filepath to an SSL cert, this will be used to enable https
	SSLCert string

	// SSLKey is a filepath to an SSL key, this will be used to enable https
	SSLKey string

	// CORSAllowedOrigin is the "Access-Control-Allow-Origin" CORS header.
	// configurable via the tegola config.toml file (set in main.go)
	CORSAllowedOrigin string = "*"

	// TileBuffer is the tile buffer to use.
	// configurable via tegola config.tomal file (set in main.go)
	TileBuffer float64 = tegola.DefaultTileBuffer
)

// NewRouter set's up the our routes.
func NewRouter(a *atlas.Atlas) *httptreemux.TreeMux {
	r := httptreemux.New()
	group := r.NewGroup("/")

	// one handler to respond to all OPTIONS requests for registered routes with our CORS headers
	r.OptionsHandler = corsHandler

	// capabilities endpoints
	group.UsingContext().Handler("GET", "/capabilities", CORSHandler(HandleCapabilities{}))
	group.UsingContext().Handler("GET", "/capabilities/:map_name", CORSHandler(HandleMapCapabilities{}))

	// map tiles
	hMapLayerZXY := HandleMapLayerZXY{Atlas: a}
	group.UsingContext().Handler("GET", "/maps/:map_name/:z/:x/:y", CORSHandler(TileCacheHandler(a, hMapLayerZXY)))
	group.UsingContext().Handler("GET", "/maps/:map_name/:layer_name/:z/:x/:y", CORSHandler(TileCacheHandler(a, hMapLayerZXY)))

	// map style
	group.UsingContext().Handler("GET", "/maps/:map_name/style.json", CORSHandler(HandleMapStyle{}))

	//	setup viewer routes, which can excluded via build flags
	setupViewer(group)

	return r
}

func New(a *atlas.Atlas, c dict.Dicter) (srv *http.Server, err error) {
	port := ":8080"
	port, err = c.String(configKeyPort, &port)
	if err != nil {
		return nil, err
	}

	sslKey := ""
	sslKey, err = c.String(configKeySSLKey, &sslKey)
	if err != nil {
		return nil, err
	}

	sslCert := ""
	sslCert, err = c.String(configKeySSLCert, &sslCert)
	if err != nil {
		return nil, err
	}



	srv = &http.Server{
		Addr: port,
	}

	if sslCert + sslKey != "" {
		cert, err := tls.LoadX509KeyPair(sslCert, sslKey)
		if err != nil {
			return nil, err
		}

		srv.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}

	}

	srv.Handler = NewRouter(a)

	return srv, nil
}

// Start starts the tile server binding to the provided port
func Start(a *atlas.Atlas, port string) *http.Server {

	// notify the user the server is starting
	log.Infof("starting tegola server on port %v", port)

	srv := &http.Server{Addr: port, Handler: NewRouter(a)}

	// start our server
	go func() {
		if SSLCert+SSLKey != "" {
			log.Error(srv.ListenAndServeTLS(SSLCert, SSLKey))
		} else {
			log.Error(srv.ListenAndServe())
		}
	}()

	return srv
}

// hostName determines the hostname:port to return based on the following hierarchy
// - HostName / Port values as configured via the config file
// - The request host / port if config HostName or Port is missing
func hostName(r *http.Request) string {
	var requestHostname string
	var requestPort string

	substrs := strings.Split(r.Host, ":")

	switch len(substrs) {
	case 1:
		requestHostname = substrs[0]
	case 2:
		requestHostname = substrs[0]
		requestPort = substrs[1]
	default:
		log.Warnf("multiple colons (':') in host string: %v", r.Host)
	}

	retHost := HostName
	if HostName == "" {
		retHost = requestHostname
	}

	if Port != "" && Port != "none" {
		return retHost + Port
	}
	if requestPort != "" && Port != "none" {
		return retHost + ":" + requestPort
	}

	return retHost
}

// various checks to determin if the request is http or https. the scheme is needed for the TileURLs
// r.URL.Scheme can be empty if a relative request is issued from the client. (i.e. GET /foo.html)
func scheme(r *http.Request) string {
	if r.Header.Get("X-Forwarded-Proto") != "" {
		return r.Header.Get("X-Forwarded-Proto")
	} else if r.TLS != nil {
		return "https"
	}

	return "http"
}

// URLRoot builds a string containing the scheme, host and port based on a combination of user defined values,
// headers and request parameters. The function is public so it can be overridden for other implementations.
var URLRoot = func(r *http.Request) string {
	return fmt.Sprintf("%v://%v", scheme(r), hostName(r))
}

// corsHanlder is used to respond to all OPTIONS requests for registered routes
func corsHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {
	w.Header().Set("Access-Control-Allow-Origin", CORSAllowedOrigin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	return
}
