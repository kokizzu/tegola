package cmd

import (
	"github.com/spf13/cobra"

	gdcmd "github.com/go-spatial/tegola/internal/cmd"
	"github.com/go-spatial/tegola/provider"
	"github.com/go-spatial/tegola/server"
	"github.com/go-spatial/tegola/internal/log"
	"github.com/go-spatial/tegola/internal/env"
)

// set by command line flag
var serverPort string

var serverCmd = &cobra.Command{
	Use:   "serve",
	Short: "Use tegola as a tile server",
	Long:  `Use tegola as a vector tile server. Maps tiles will be served at /maps/:map_name/:z/:x/:y`,
	Run: func(cmd *cobra.Command, args []string) {
		gdcmd.New()
		initConfig()
		gdcmd.OnComplete(provider.Cleanup)

		if serverPort != "" {
			conf.Webserver.Port = env.StringPtr(env.String(serverPort))
		}

		// set our server version
		server.Version = Version

		// start our webserver
		srv, err := server.New(nil, conf.Webserver)
		if err != nil {
			log.Fatal(err)
		}

		shutdown(srv)
		<-gdcmd.Cancelled()
		gdcmd.Complete()

	},
}
