package svg

import (
	"encoding/xml"
	"fmt"

	svg "github.com/ajstarks/svgo"
)

const DefaultSpacing = 10

type Canvas struct {
	*svg.SVG
	Board  MinMax
	Region MinMax
}

func (canvas *Canvas) DrawPoint(x, y int, fill string) {
	canvas.Gstyle("text-anchor:middle;font-size:8;fill:white;stroke:black")
	canvas.Circle(x, y, 1, "fill:"+fill)
	canvas.Text(x, y+5, fmt.Sprintf("(%v %v)", x, y))
	canvas.Gend()
}

func drawGrid(canvas *Canvas, mm *MinMax, n int, label bool, id, style, pointstyle, ostyle string) {
	x, y, w, h := int(mm.MinX), int(mm.MinY), int(mm.Width()), int(mm.Height())
	canvas.Group(fmt.Sprintf(`id="%v"`, id), fmt.Sprintf(`style="%v"`, style))
	// Draw all the horizontal and vertical lines.
	for i := x; i < x+w; i += n {
		canvas.Line(i, y, i, y+h)
	}
	for i := y; i < y+h; i += n {
		canvas.Line(x, i, x+w, i)
	}
	if !label {
		canvas.Gend()
		return
	}
	canvas.Gend()
	canvas.Group(fmt.Sprintf(`id="%v"`, id+"_origin"), fmt.Sprintf(`style="%v"`, ostyle))
	canvas.Line(0, y, 0, h)
	canvas.Line(x, 0, w, 0)
	canvas.Gend()
	canvas.Group(fmt.Sprintf(`id="%v"`, id+"_points"), fmt.Sprintf(`style="%v"`, pointstyle))
	for i := x; i < w; i += n {
		canvas.Circle(i, y, 2, "fill:black")
		canvas.Text(i, y-5, fmt.Sprintf("% 3v", i))
	}
	for i := x; i < h; i += n {
		canvas.Circle(x, i, 2, "fill:black")
		canvas.Text(x-10, i+2, fmt.Sprintf("% 3v", i))
	}
	canvas.Gend()
}

func (canvas *Canvas) Comment(s string) *Canvas {
	fmt.Fprint(canvas.Writer, "<!-- \n")
	xml.Escape(canvas.Writer, []byte(s))
	fmt.Fprint(canvas.Writer, "\n -->")
	return canvas
}
func (canvas *Canvas) Commentf(format string, a ...interface{}) *Canvas {
	fmt.Fprint(canvas.Writer, "<!-- \n")
	xml.Escape(canvas.Writer, []byte(fmt.Sprintf(format, a...)))
	fmt.Fprint(canvas.Writer, "\n -->")
	return canvas
}

func (canvas *Canvas) DrawGrid(n int, label bool, style string) {
	drawGrid(canvas, &canvas.Board, n, label, fmt.Sprintf("board_%v", n), style, "text-anchor:middle;font-size:8;fill:white;stroke:black", "stroke:black")
}

func (canvas *Canvas) DrawRegion(withGrid bool) {

	canvas.Group(`id="region"`, `style="opacity:0.2"`)
	canvas.Rect(int(canvas.Region.MinX), int(canvas.Region.MinY), int(canvas.Region.Width()), int(canvas.Region.Height()), "stroke-dasharray:5,5;fill:red;opacity:0.3;"+fmt.Sprintf(";stroke:rgb(%v,%v,%v)", 0, 200, 0))

	if withGrid {
		drawGrid(canvas, &canvas.Region, 10, false, "region_10", "stroke:red;opacity:0.2", "text-anchor:middle;font-size:8;fill:white;stroke:red", "stroke:red")
		drawGrid(canvas, &canvas.Region, 100, true, "region_100", "stroke:red;opacity:0.3", "text-anchor:middle;font-size:8;fill:white;stroke:red", "stroke:red")
	}
	canvas.Gend()
}

/*
func (canvas *Canvas) DrawPolygon(p geom.Polygon, id string, style string) int {
	var points []maths.Pt
	canvas.Group(`id="`+id+`"`, `style="opacity:1"`)
	canvas.Gid("polygon_path")
	path := ""
	pointCount := 0
	for _, l := range p {
		idx := len(l)
		if idx <= 0 {
			continue
		}
		for i, pt := range l {
			points = append(points, maths.Pt{X: pt.X(), Y: pt.Y()})
			if i == 0 {
				path += "M "
			} else {
				path += "L "
			}
			path += fmt.Sprintf("%v %v ", pt.X(), pt.Y())
			pointCount++
		}

		path += "Z "
	}
	canvas.Commentf("Point Count: %v", pointCount)
	canvas.Path(path, fmt.Sprintf(`id="%v_%v"`, id, pointCount), style)
	canvas.Gend()
	canvas.Gend()
	return pointCount
}

func (canvas *Canvas) DrawMultiPolygon(mp geom.MultiPolygon, id string, style string) int {
	canvas.Gid(id)
	count := 0
	for i, p := range mp.Polygons() {
		count += canvas.DrawPolygon(geom.Polygon(p), fmt.Sprintf("%v_mp_%v", id, i), style)
	}
	canvas.Gend()
	return count
}

func (canvas *Canvas) DrawLine(l geom.LineString, id string, style string) {

	canvas.Gid(id)
	path := ""
	for i, pt := range l {
		if i == 0 {
			path += "M "
		} else {
			path += "L "
		}
		path += fmt.Sprintf("%v %v ", pt[0], pt[1])
	}
	canvas.Path(path, style)
	canvas.Gend()
}

func (canvas *Canvas) DrawMathSegments(ls []maths.Line, s ...string) {
	log.Printf("Drawing lines(%v) ", len(ls))
	for _, line := range ls {
		canvas.Line(
			int(line[0].X),
			int(line[0].Y),
			int(line[1].X),
			int(line[1].Y),
			s...,
		)
	}
}
func (canvas *Canvas) DrawMathPoints(pts []maths.Pt, s ...string) {
	log.Printf("Drawing Points (%v)", len(pts))
	prefix := "M"
	var path string
	for i := range pts {
		path += fmt.Sprintf("%v %v %v ", prefix, pts[i].X, pts[i].Y)
		prefix = "L"
	}
	canvas.Path(path, s...)
}

func (canvas *Canvas) DrawMultiLine(ml geom.MultiLine, id string, style string) {
	canvas.Gid(id)
	for i, l := range ml.LineStrings() {
		canvas.DrawLine(l, fmt.Sprintf("%v_%v", id, i), style)
	}
	canvas.Gend()
}

func (canvas *Canvas) DrawGeometry(geo tegola.Geometry, id string, style string, pointStyle string, drawPoints bool) int {
	count := 0
	switch g := geo.(type) {
	case geom.MultiLine:
		canvas.DrawMultiLine(g, "multiline_"+id, style)
	case geom.MultiPolygon:
		count += canvas.DrawMultiPolygon(g, "multipolygon_"+id, style, pointStyle)
	case geom.Polygon:
		count += canvas.DrawPolygon(g, "polygon_"+id, style, pointStyle, drawPoints)
	case geom.LineString:
		canvas.DrawLine(g, "line_"+id, style)
	case geom.Point:
		canvas.Gid("point_" + id)
		canvas.DrawPoint(int(g[0]), int(g[1]), pointStyle)
		canvas.Gend()
	case geom.MultiPoint:
		canvas.Gid("multipoint_" + id)
		for i, p := range g.Points() {
			canvas.Gid(fmt.Sprintf("mp_%v", i))
			canvas.DrawPoint(int(p.X()), int(p.Y()), pointStyle)
			canvas.Gend()
		}
		canvas.Gend()
	}
	return count
}

func (canvas *Canvas) Init(writer io.Writer, w, h int, grid bool) *Canvas {
	if canvas == nil {
		panic("Canvas can not be nil!")
	}
	canvas.SVG = svg.New(writer)

	canvas.Startview(w, h, int(canvas.Board.MinX-20), int(canvas.Board.MinY-20), int(canvas.Board.MaxX+20), int(canvas.Board.MaxY+20))
	if grid {
		canvas.GroupFn([]string{
			`id="grid"`,
		}, func(canvas *Canvas) {
			canvas.DrawGrid(10, false, "stroke:gray")
			canvas.DrawGrid(100, true, "stroke:black")
		},
		)

	}
	return canvas
}

func (canvas *Canvas) GroupFn(attr []string, fn func(c *Canvas)) {
	canvas.SVG.Group(attr...)
	fn(canvas)
	canvas.SVG.Gend()
}
*/
