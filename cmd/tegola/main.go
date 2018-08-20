package main

import (
	"fmt"
	"log"
	"os"

	_ "github.com/theckman/goconstraint/go1.8/gte"

	"github.com/go-spatial/tegola/cmd/tegola/cmd"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
