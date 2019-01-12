package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	svg "github.com/ajstarks/svgo"
	"github.com/go-spatial/cobra"
	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/encoding/wkb"
)

var wkbCmd = &cobra.Command{
	Use:   "wkb",
	Short: "Draw the given wkb file",
	Long:  "The wkb command will draw out feature and the various stages of the encoding process.",
	Run:   wkbCommand,
}

var wkbInputFilename string
var wkbOutputFilename string

func init() {
	wkbCmd.Flags().StringVarP(&wkbInputFilename, "input", "i", "", "The file to draw out.")
	wkbCmd.Flags().StringVarP(&wkbOutputFilename, "output", "o", "", "the output file. If not provided it will be name of the input with svg ext.")
}

func wkbCommand(cmd *cobra.Command, args []string) {
	if wkbInputFilename == "" {
		fmt.Fprintf(os.Stderr, "Need an input file.")
		os.Exit(1)
	}
	if wkbOutputFilename == "" {
		if ext := filepath.Ext(wkbInputFilename); len(ext) != 0 {
			wkbOutputFilename = wkbInputFilename[:len(wkbInputFilename)-len(ext)] + ".svg"
		} else {
			wkbOutputFilename = wkbInputFilename + ".svg"
		}
	}

	// Let's load the wkb file.
	file, err := os.Open(wkbInputFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unabled to open %v : %v", wkbInputFilename, err)
		os.Exit(1)
	}

	defer file.Close()
	geo, err := wkb.Decode(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unabled to decode %v : %v", wkbInputFilename, err)
		os.Exit(1)
	}

	switch geo.(type) {
	case geom.Polygoner, geom.MultiPolygoner:
	default:
		fmt.Fprintf(os.Stderr, "Only support Polygon and MultiPolygon, got type %T: %v ", geo, wkbInputFilename)
		os.Exit(1)
	}

	clipbox, err := geom.NewExtentFromGeometry(geo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unabled to get the extent for geomentry: %v :%v", geo, err)
		os.Exit(1)
	}

	final, oTriangles, iTriangles, err := GetParts(clipbox, geo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unabled to get parts for %v : %v", wkbInputFilename, err)
		os.Exit(1)
	}

	ofile, err := os.Create(wkbOutputFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unabled to create output file %v: %v", wkbOutputFilename, err)
		os.Exit(1)
	}
	defer ofile.Close()

	// Draw out the svg file.
	canvas := svg.New(ofile)
	canvas.Startraw()
	DrawParts(canvas, clipbox, geo, final, oTriangles, iTriangles)
	canvas.End()

}
