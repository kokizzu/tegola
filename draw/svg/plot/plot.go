package plot

import (
	"fmt"
	"math"
	"strconv"

	svg "github.com/ajstarks/svgo"
	"github.com/go-spatial/geom"
)

type grid struct {
	xscale      float64
	xoffset     float64
	yscale      float64
	yoffset     float64
	pointRadius int
	pointStyle  string
	c           *svg.SVG
}

const (
	DefaultPointRadius = 5
)

func NewDefaultUniformGrid(canvas *svg.SVG, xmin, ymin float64) *grid {
	return NewGrid(canvas, xmin, ymin, 100, 100, 100, 100)
}
func NewUniformGrid(canvas *svg.SVG, xmin, ymin, start, scale float64) *grid {
	return NewGrid(canvas, xmin, ymin, start, scale, start, scale)
}
func NewGrid(canvas *svg.SVG, xmin, ymin float64, xstart, xscale, ystart, yscale float64) *grid {

	return &grid{
		xscale:      xscale,
		xoffset:     xstart - (xmin * xscale),
		yscale:      yscale,
		yoffset:     ystart - (ymin * yscale),
		pointRadius: DefaultPointRadius,
		pointStyle:  `style="fill:#e6f7ff;stroke-width:0px; stroke:#e6f7ff;"`,
		c:           canvas,
	}
}

func (g *grid) SeperateXY(pts ...[2]float64) (xs []int, ys []int) {
	xs = make([]int, len(pts))
	ys = make([]int, len(pts))
	for i, pt := range pts {
		xs[i], ys[i] = g.X(pt[0]), g.Y(pt[1])
	}
	return xs, ys
}

func (g *grid) Point(x, y float64) (int, int) { return g.X(x), g.Y(y) }

func (g *grid) X(x float64) int { return int((x * g.xscale) + g.xoffset) }
func (g *grid) Y(y float64) int { return int((y * g.yscale) + g.yoffset) }

// LineString will use a Polygon to draw out the linestring.
func (g *grid) LineString(ls [][2]float64, s ...string) {
	xs, ys := g.SeperateXY(ls...)
	g.c.Polygon(xs, ys, s...)
}

func (g *grid) Geometry(id string, geo geom.Geometry, s ...string) {
	switch gg := geo.(type) {
	case geom.Pointer:
		xy := gg.XY()
		ss := append([]string{fmt.Sprintf(`id="%v" class="point"`, id)}, s...)
		g.c.Circle(g.X(xy[0]), g.Y(xy[1]), DefaultPointRadius, ss...)
	case geom.MultiPointer:
		g.Gid(id, func() {
			pts := gg.Points()
			for _, pt := range pts {
				ss := append([]string{`class="point"`}, s...)
				g.c.Circle(g.X(pt[0]), g.Y(pt[1]), DefaultPointRadius, ss...)
			}
		})
	case geom.LineStringer:
		pts := gg.Verticies()
		if len(pts) <= 0 {
			return
		}

		g.LineString(gg.Verticies(), append([]string{fmt.Sprintf(`id="%v" class="linestring"`, id)}, s...)...)

	case geom.MultiLineStringer:
		lines := gg.LineStrings()
		g.Gid(id, func() {
			for _, line := range lines {
				g.LineString(line, append([]string{fmt.Sprintf(`id="%v" class="linestring"`, id)}, s...)...)
			}
		})

	case geom.Polygoner:
		rings := gg.LinearRings()
		g.Gid(id, func() {
			if len(rings) <= 0 {
				return
			}
			g.LineString(rings[0], append([]string{`class="main-linearring,linearring"`}, s...)...)
			for i := 1; i < len(rings); i++ {
				g.LineString(rings[i], append([]string{`class="cutout-linearring,linearring"`}, s...)...)
			}
		})
	case geom.MultiPolygoner:
		polygons := gg.Polygons()
		g.Gid(id, func() {
			if len(polygons) <= 0 {
				return
			}
			for i, ply := range polygons {
				g.Gid(fmt.Sprintf("%v_%v", id, i), func() {
					if len(ply) <= 0 {
						return
					}
					g.LineString(ply[0], append([]string{`class="main-linearring, linearring"`}, s...)...)
					for i := 1; i < len(ply); i++ {
						g.LineString(ply[i], append([]string{`class="cutout-linearring,linearring"`}, s...)...)
					}
				})
			}
		})

	case geom.Collectioner:
		geoms := gg.Geometries()
		g.Gid(id, func() {
			if len(geoms) <= 0 {
				return
			}
			for i := range geoms {
				g.Geometry(fmt.Sprintf("%v_%v", id, i), geoms[i], s...)
			}
		})

	case geom.Circle:
		gr := g.X(gg.Center[0]+math.Abs(gg.Radius)) - g.X(gg.Center[0])
		ss := append([]string{fmt.Sprintf(`id="%v" class="circle"`, id)}, s...)
		g.c.Circle(g.X(gg.Center[0]), g.Y(gg.Center[1]), gr, ss...)

	case geom.Line:
		ss := append([]string{fmt.Sprintf(`id="%v" class="line"`, id)}, s...)
		g.c.Line(g.X(gg[0][0]), g.Y(gg[0][1]), g.X(gg[1][0]), g.Y(gg[1][1]), ss...)

	case geom.Triangle:
		ss := append([]string{fmt.Sprintf(`id="%v" class="triangle"`, id)}, s...)
		g.LineString(gg[:], ss...)

	case geom.MinMaxer:
		minx, maxx := g.X(gg.MinX()), g.X(gg.MaxX())
		miny, maxy := g.Y(gg.MinY()), g.Y(gg.MaxY())
		g.c.Rect(minx, miny, maxx-minx, maxy-miny, s...)

	}

}

func (g *grid) Gid(gid string, blk func()) {
	g.c.Gid(gid)
	blk()
	g.c.Gend()
}

func (g *grid) Def(blk func()) {
	g.c.Def()
	blk()
	g.c.DefEnd()
}

func (g *grid) DrawGridAndAxis(Minx, Miny, Maxx, Maxy float64, drawGrid bool) {
	const (
		sqrWidth  = 10
		sqrHeight = 10
	)

	minx, maxx := g.X(Minx), g.X(Maxx)
	miny, maxy := g.Y(Miny), g.Y(Maxy)

	g.Gid("grid", func() {

		if drawGrid {
			// Set up a row of dots.
			g.Gid("gridHorizLines", func() {
				for y := Miny; y < Maxy; y += 1.0 {
					gy := g.Y(y)
					g.c.Line(minx, gy, maxx, gy, g.pointStyle)
				}
			})
			g.Gid("gridVerticalLines", func() {
				for x := Minx; x < Maxx; x += 1.0 {
					gx := g.X(x)
					g.c.Line(gx, miny, gx, maxy, g.pointStyle)
				}
			})

		}
		g.Gid("gridAxis", func() {
			// x-axis
			g.c.Line(g.X(Minx-1), miny, g.X(Maxx+1), miny, "id=\"xaxis\"")
			g.c.Line(g.X(Minx-1), maxy, g.X(Maxx+1), maxy, "id=\"xaxis-end\"")
			for x := int(Minx + 1); x < int(Maxx); x++ {
				strx := strconv.Itoa(x)
				yscale := int(g.yscale * .10)
				g.c.Text(g.X(float64(x)), miny-yscale, strx)
				g.c.Text(g.X(float64(x)), maxy+yscale, strx)
			}
			// y-Axis
			g.c.Line(minx, g.Y(Miny-1), minx, g.Y(Maxy+1), "id=\"yaxis\"")
			g.c.Line(maxx, g.Y(Miny-1), maxx, g.Y(Maxy+1), "id=\"yaxis-end\"")
			nfmt := fmt.Sprintf("%% %vv", len(fmt.Sprintf("%v", int(Maxy))))
			for y := int(Miny + 1); y < int(Maxy); y++ {
				stry := fmt.Sprintf(nfmt, y)
				g.c.Text(minx-10, g.Y(float64(y)), stry)
				g.c.Text(maxx+10, g.Y(float64(y)), stry)
			}
		})
	})
}
