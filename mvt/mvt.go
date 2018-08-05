package mvt

import (
	"log"
	"os"
	"strconv"
	"strings"
)

var (
	simplifyGeometries    = true
	simplificationMaxZoom = 10
)

func init() {
	options := strings.ToLower(os.Getenv("TEGOLA_OPTIONS"))
	if strings.Contains(options, "dontsimplifygeo") {
		simplifyGeometries = false
		log.Println("simplification of geometries is off")
	}

	if strings.Contains(options, "simplifymaxzoom=") {
		idx := strings.Index(options, "simplifymaxzoom=")
		idx += 16
		eidx := strings.IndexAny(options[idx:], ",.\t \n")

		if eidx == -1 {
			eidx = len(options)
		} else {
			eidx += idx
		}

		i, err := strconv.Atoi(options[idx:eidx])
		if err != nil {
			log.Printf("did not understand the value (%v) for SimplifyMaxZoom. using default (%v).", options[idx:eidx], simplificationMaxZoom)
			return
		}

		simplificationMaxZoom = int(i + 1)

		log.Printf("setting SimplifyMaxZoom to %v", int(i))
	}
}
