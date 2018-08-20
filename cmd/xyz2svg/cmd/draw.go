package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	svg "github.com/ajstarks/svgo"
	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/planar"
	"github.com/go-spatial/geom/planar/makevalid"
	"github.com/go-spatial/geom/planar/makevalid/hitmap"
	"github.com/go-spatial/geom/planar/makevalid/walker"
	"github.com/go-spatial/geom/slippy"
	"github.com/go-spatial/tegola"
	"github.com/go-spatial/tegola/cmd/internal/register"
	"github.com/go-spatial/tegola/config"
	"github.com/go-spatial/tegola/dict"
	"github.com/go-spatial/tegola/draw/svg/plot"
	"github.com/go-spatial/tegola/mvt"
	"github.com/go-spatial/tegola/provider"
)

const GridStyle = `
g#grid {
	stroke-width: 1px;
	stroke: #e6f7ff;
}
g#grid line {
	stroke-width: 1px;
	stroke: #e6f7ff;
}
g#gridPoints circle {
	fill: #e6f7ff;
	stroke-width: 0px;
	stroke: #e6f7ff;
}
g#grid text {
	font:  200px helvetica; fill: #01AAF9;
	alignment-baseline: middle;
}

g#ringPoints circle.constrained {
	stroke:black;
	stroke-width:3px;
	fill: black;
}
g#ringPoints circle.unconstrained {
	stroke:black;
	stroke-width:3px;
	fill: none;
}

#markerHalfTopArrow {
	fill: gray;
}
`

var drawCmd = &cobra.Command{
	Use:   "draw",
	Short: "Draw the requested tile or feature",
	Long:  "The draw command will draw out the feature and the various stages of the encoding process.",
	Run:   drawCommand,
}

var drawOutputBaseDir string
var drawOutputFilenameFormat string

func init() {
	drawCmd.Flags().StringVarP(&drawOutputBaseDir, "output", "o", "_svg_files", "Directory to write svg files to.")
	drawCmd.Flags().StringVarP(&drawOutputFilenameFormat, "format", "f", "{{base_dir}}/z{{z}}_x{{x}}_y{{y}}/{{layer_name}}/geo_{{gid}}_{{count}}.{{ext}}", "filename format")
}

type drawFilename struct {
	z, x, y uint
	basedir string
	format  string
	ext     string
}

func (dfn drawFilename) insureFilename(provider string, layer string, gid int, count int) (string, error) {
	r := strings.NewReplacer(
		"{{base_dir}}", dfn.basedir,
		"{{ext}}", dfn.ext,
		"{{layer_name}}", layer,
		"{{provider_name}}", provider,
		"{{gid}}", strconv.FormatInt(int64(gid), 10),
		"{{count}}", strconv.FormatInt(int64(count), 10),
		"{{z}}", strconv.FormatInt(int64(dfn.z), 10),
		"{{x}}", strconv.FormatInt(int64(dfn.x), 10),
		"{{y}}", strconv.FormatInt(int64(dfn.y), 10),
	)
	filename := filepath.Clean(r.Replace(dfn.format))
	basedir := filepath.Dir(filename)
	if err := os.MkdirAll(basedir, 0711); err != nil {
		return "", err
	}
	return filename, nil
}

func (dfn drawFilename) createFile(provider string, layer string, gid int, count int) (string, *os.File, error) {
	fname, err := dfn.insureFilename(provider, layer, gid, count)
	if err != nil {
		return "", nil, err
	}
	file, err := os.Create(fname)
	return fname, file, err
}

func drawCommand(cmd *cobra.Command, args []string) {

	z, x, y, err := parseTileString(zxystr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid zxy (%v): %v\n", zxystr, err)
		os.Exit(1)
	}

	config, err := config.LoadAndValidate(configFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid config (%v): %v\n", configFilename, err)
		os.Exit(1)
	}
	dfn := drawFilename{
		z:       z,
		x:       x,
		y:       y,
		ext:     "svg",
		format:  drawOutputFilenameFormat,
		basedir: drawOutputBaseDir,
	}

	// convert []env.Map -> []dict.Dicter
	provArr := make([]dict.Dicter, len(config.Providers))
	for i := range provArr {
		provArr[i] = config.Providers[i]
	}

	// register providers
	providers, err := register.Providers(provArr)
	if err != nil {
		log.Fatalf("Error loading providers in config(%v): %v\n", configFilename, err)
	}

	prv, lyr := splitProviderLayer(providerString)
	var allprvs []string
	for name := range providers {
		allprvs = append(allprvs, name)
	}
	var prvs = []string{prv}
	// If prv is "" we are going to go through every feature.
	if prv == "" {
		prvs = allprvs
	}
	for _, name := range prvs {
		tiler, ok := providers[name]
		if !ok {
			fmt.Fprintf(os.Stderr, "Skipping  did not find provider %v\n", name)
			fmt.Fprintf(os.Stderr, "known providers: %v\n", strings.Join(allprvs, ", "))
			continue
		}
		var layers []string
		if lyr == "" {
			lysi, err := tiler.Layers()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Skipping error getting layers for provider (%v): %v\n", name, err)
			}
			for _, li := range lysi {
				layers = append(layers, li.Name())
			}
		} else {
			layers = append(layers, lyr)
		}
		slippyTile := slippy.NewTile(dfn.z, dfn.x, dfn.y, tegola.DefaultTileBuffer, tegola.WebMercator)
		for _, layerName := range layers {

			if err := tiler.TileFeatures(context.Background(),
				layerName,
				slippyTile,
				dfn.DrawFeature(name, layerName, gid),
			); err != nil {
				panic(err)
			}
		}
	}
	provider.Cleanup()
}

func DrawParts(canvas *svg.SVG, clipbox *geom.Extent, ori, final geom.Geometry, outsideTriangles, insideTriangles []geom.Triangle) {

	grid := plot.NewUniformGrid(canvas, clipbox.MinX(), clipbox.MinY(), 100, DrawingScale)
	grid.Def(func() {
		canvas.Style("text/css", GridStyle)
		//grid.DrawGridAndAxis(clipbox.MinX(), clipbox.MinY(), clipbox.MaxX(), clipbox.MaxY(), true)
	})
	canvas.Use(0, 0, "grid")

	grid.Gid("outside_triangles", func() {
		for i, triangle := range outsideTriangles {
			grid.Geometry(fmt.Sprintf("otri_%v", i), triangle, `style="fill: gray; stroke:black; stroke-width:1px"`)
			cpt := geom.Point(triangle.Center())
			grid.Geometry(fmt.Sprintf("otri_cpt_%v", i), cpt, `style="fill: white;stroke:white; stroke-width:6px"`)
		}
	})
	grid.Gid("inside_triangles", func() {
		for i, triangle := range insideTriangles {
			grid.Geometry(fmt.Sprintf("itri_%v", i), triangle, `style="fill: green; stroke:black; stroke-width:6px"`)
			cpt := geom.Point(triangle.Center())
			grid.Geometry(fmt.Sprintf("itri_cpt_%v", i), cpt, `style="fill: white;stroke:white; stroke-width:6px"`)

		}
	})
	grid.Geometry("clipbox", clipbox, `style="fill:red; fill-opacity:0.4; stroke:red; stroke-width:6px"`)
	grid.Geometry("simplified_polygon", ori, `style="fill:yellow; fill-opacity:0.4"`)
	grid.Geometry("generate_polygon", final, `style="fill:blue;fill-opacity:0.3"`)
}

func GetParts(clipbox *geom.Extent, geo geom.Geometry) (final geom.Geometry, outsideTriangles, insideTriangles []geom.Triangle, err error) {
	var mply geom.MultiPolygon
	ctx := context.Background()

	hm := hitmap.MustNew(nil, geo)
	switch gg := geo.(type) {
	case geom.Polygon:
		mply = geom.MultiPolygon{[][][2]float64(gg)}
	case geom.MultiPolygon:
		mply = gg
	default:
		return nil, nil, nil, fmt.Errorf("Only support Polygoner and MultiPolygon, got type %T", geo)
	}

	segments, err := makevalid.Destructure(ctx, clipbox, &mply)
	if err != nil {
		return nil, nil, nil, err
	}
	if len(segments) == 0 {
		log.Printf("# clipped the multipolygon out of existance.")
		return nil, nil, nil, err
	}

	/* Let's draw out all the triangles. which means we have to do the triangulation twice. */
	allTriangles, err := makevalid.TriangulateGeometry(ctx, segments)
	if err != nil {
		return nil, nil, nil, err
	}

	for i, triangle := range allTriangles {
		if hm.LabelFor(triangle.Center()) == planar.Outside {
			outsideTriangles = append(outsideTriangles, allTriangles[i])
		} else {
			insideTriangles = append(insideTriangles, allTriangles[i])
		}
	}

	fixedMultiPolygon := walker.MultiPolygon(ctx, insideTriangles)
	return fixedMultiPolygon, outsideTriangles, insideTriangles, nil
}

func (dfn drawFilename) DrawFeature(pname, name string, gid int) func(f *provider.Feature) error {

	ttile := tegola.NewTile(dfn.z, dfn.x, dfn.y)
	cursor := mvt.NewCursor(ttile)

	pbb, err := ttile.PixelBufferedBounds()
	if err != nil {
		panic(err)
	}

	count := 0
	clipbox := geom.NewExtent([2]float64{pbb[0], pbb[1]}, [2]float64{pbb[2], pbb[3]})

	return func(f *provider.Feature) error {
		if gid != -1 && f.ID != uint64(gid) {
			// Skip the feature.
			return nil
		}

		switch f.Geometry.(type) {
		case geom.Polygoner:
		case geom.MultiPolygoner:
		default:
			panic(fmt.Sprintf("Only support Polygoner and MultiPolygon, got type %T for %v", f.Geometry, f.ID))
		}

		// Scale
		g, err := cursor.ProjectGeometry(f.Geometry)
		if err != nil {
			return err
		}

		// Simplify
		sg, err := mvt.SimplifyGeometryGeom(context.Background(), g, ttile.ZEpislon())
		if err != nil {
			return err
		}

		count++
		finalgeom, oTriangles, iTriangles, err := GetParts(clipbox, sg)
		if err != nil {
			return err
		}

		ffname, writer, err := dfn.createFile(pname, name, gid, count)
		if err != nil {
			return err
		}
		log.Printf("Writing to file: %v\n", ffname)

		// Draw out the svg file.
		canvas := svg.New(writer)
		canvas.Startraw()
		DrawParts(canvas, clipbox, sg, finalgeom, oTriangles, iTriangles)
		canvas.End()
		writer.Close()

		return nil
	}
}
