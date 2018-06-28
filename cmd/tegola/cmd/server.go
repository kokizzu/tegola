package cmd

import (
	"github.com/spf13/cobra"

	gdcmd "github.com/go-spatial/tegola/internal/cmd"
	"github.com/go-spatial/tegola/provider"
	"github.com/go-spatial/tegola/server"
	"github.com/go-spatial/tegola/internal/log"
)

var (
	serverPort      string
	defaultHTTPPort = ":8080"
)

var serverCmd = &cobra.Command{
	Use:   "serve",
	Short: "Use tegola as a tile server",
	Long:  `Use tegola as a vector tile server. Maps tiles will be served at /maps/:map_name/:z/:x/:y`,
	Run: func(cmd *cobra.Command, args []string) {
		gdcmd.New()
		initConfig()
		gdcmd.OnComplete(provider.Cleanup)

		// check config for server port setting
		// if you set the port via the comand line it will override the port setting in the config
		if serverPort == defaultHTTPPort && conf.Webserver.Port != "" {
			serverPort = string(conf.Webserver.Port)
		}

		// set our server version
		server.Version = Version
		server.HostName = string(conf.Webserver.HostName)

		// set the CORSAllowedOrigin if a value is provided
		if conf.Webserver.CORSAllowedOrigin != "" {
			server.CORSAllowedOrigin = string(conf.Webserver.CORSAllowedOrigin)
		}

		// set tile buffer
		if conf.TileBuffer > 0 {
			server.TileBuffer = float64(conf.TileBuffer)
		}

		if conf.Webserver.SSLCert+conf.Webserver.SSLKey != "" {
			if conf.Webserver.SSLCert == "" {
				// error
				log.Fatal("config must have both or nether ssl_key and ssl_cert, missing ssl_cert")
			}

			if conf.Webserver.SSLKey == "" {
				// error
				log.Fatal("config must have both or nether ssl_key and ssl_cert, missing ssl_key")
			}

			server.SSLCert = string(conf.Webserver.SSLCert)
			server.SSLKey = string(conf.Webserver.SSLKey)
		}

		// start our webserver
		srv := server.Start(nil, serverPort)
		shutdown(srv)
		<-gdcmd.Cancelled()
		gdcmd.Complete()

	},
}
