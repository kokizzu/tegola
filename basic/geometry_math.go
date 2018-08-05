package basic

import (
	"fmt"

	"github.com/go-spatial/tegola"
	"github.com/go-spatial/tegola/maths/webmercator"
)

// ApplyToPoints applys the given function to each point in the geometry and any sub geometries, return a new transformed geometry.
func ApplyToPoints(geometry tegola.Geometry, f func(coords ...float64) ([]float64, error)) (G, error) {
	switch geo := geometry.(type) {
	default:
		return G{}, fmt.Errorf("Unknown Geometry: %+v", geometry)
	case tegola.Point:
		c, err := f(geo.X(), geo.Y())
		if err != nil {
			return G{}, err
		}
		if len(c) < 2 {
			return G{}, fmt.Errorf("Function did not return minimum number of coordinates got %v expected 2", len(c))
		}
		return G{Point{c[0], c[1]}}, nil
	case tegola.Point3:
		c, err := f(geo.X(), geo.Y(), geo.Z())
		if err != nil {
			return G{}, err
		}
		if len(c) < 3 {
			return G{}, fmt.Errorf("Function did not return minimum number of coordinates got %v expected 3", len(c))
		}
		return G{Point3{c[0], c[1], c[2]}}, nil
	case tegola.MultiPoint:
		var pts MultiPoint
		for _, pt := range geo.Points() {
			c, err := f(pt.X(), pt.Y())
			if err != nil {
				return G{}, err
			}
			if len(c) < 2 {
				return G{}, fmt.Errorf("Function did not return minimum number of coordinates got %v expected 2", len(c))
			}
			pts = append(pts, Point{c[0], c[1]})
		}
		return G{pts}, nil
	case tegola.LineString:
		var line Line
		for _, ptGeo := range geo.Subpoints() {
			c, err := f(ptGeo.X(), ptGeo.Y())
			if err != nil {
				return G{}, err
			}
			if len(c) < 2 {
				return G{}, fmt.Errorf("Function did not return minimum number of coordinates got %v expected 2", len(c))
			}
			line = append(line, Point{c[0], c[1]})
		}
		return G{line}, nil
	case tegola.MultiLine:
		var line MultiLine
		for i, lineGeo := range geo.Lines() {
			geoLine, err := ApplyToPoints(lineGeo, f)
			if err != nil {
				return G{}, fmt.Errorf("Got error converting line(%v) of multiline: %v", i, err)
			}
			if !geoLine.IsLine() {
				panic("We did not get the conversion we were expecting")
			}
			line = append(line, geoLine.AsLine())
		}
		return G{line}, nil
	case tegola.Polygon:
		var poly Polygon
		for i, line := range geo.Sublines() {
			geoLine, err := ApplyToPoints(line, f)
			if err != nil {
				return G{}, fmt.Errorf("Got error converting line(%v) of polygon: %v", i, err)
			}
			poly = append(poly, geoLine.AsLine())
		}
		return G{poly}, nil
	case tegola.MultiPolygon:
		var mpoly MultiPolygon
		for i, poly := range geo.Polygons() {
			geoPoly, err := ApplyToPoints(poly, f)
			if err != nil {
				return G{}, fmt.Errorf("Got error converting poly(%v) of multipolygon: %v", i, err)
			}
			mpoly = append(mpoly, geoPoly.AsPolygon())
		}
		return G{mpoly}, nil
	}
}

// CloneGeomtry returns a deep clone of the Geometry.
func CloneGeometry(geometry tegola.Geometry) (G, error) {
	switch geo := geometry.(type) {
	default:
		return G{}, fmt.Errorf("Unknown Geometry: %+v", geometry)
	case tegola.Point:
		return G{Point{geo.X(), geo.Y()}}, nil
	case tegola.Point3:
		return G{Point3{geo.X(), geo.Y(), geo.Z()}}, nil
	case tegola.MultiPoint:
		var pts MultiPoint
		for _, pt := range geo.Points() {
			pts = append(pts, Point{pt.X(), pt.Y()})
		}
		return G{pts}, nil
	case tegola.LineString:
		var line Line
		for _, ptGeo := range geo.Subpoints() {
			line = append(line, Point{ptGeo.X(), ptGeo.Y()})
		}
		return G{line}, nil
	case tegola.MultiLine:
		var line MultiLine
		for i, lineGeo := range geo.Lines() {
			geom, err := CloneGeometry(lineGeo)
			if err != nil {
				return G{}, fmt.Errorf("Got error converting line(%v) of multiline: %v", i, err)
			}
			line = append(line, geom.AsLine())
		}
		return G{line}, nil
	case tegola.Polygon:
		var poly Polygon
		for i, line := range geo.Sublines() {
			geom, err := CloneGeometry(line)
			if err != nil {
				return G{}, fmt.Errorf("Got error converting line(%v) of polygon: %v", i, err)
			}
			poly = append(poly, geom.AsLine())
		}
		return G{poly}, nil
	case tegola.MultiPolygon:
		var mpoly MultiPolygon
		for i, poly := range geo.Polygons() {
			geom, err := CloneGeometry(poly)
			if err != nil {
				return G{}, fmt.Errorf("Got error converting poly(%v) of multipolygon: %v", i, err)
			}
			mpoly = append(mpoly, geom.AsPolygon())
		}
		return G{mpoly}, nil
	}
}

// ToWebMercator takes a SRID and a geometry encode using that srid, and returns a geometry encoded as a WebMercator.
func ToWebMercator(SRID uint64, geometry tegola.Geometry) (G, error) {
	switch SRID {
	default:
		return G{}, fmt.Errorf("Don't know how to convert from %v to %v.", tegola.WebMercator, SRID)
	case tegola.WebMercator:
		// Instead of just returning the geometry, we are cloning it so that the user of the API can rely
		// on the result to alway be a copy. Instead of being a reference in the on instance that it's already
		// in the same SRID.

		return CloneGeometry(geometry)
	case tegola.WGS84:
		return ApplyToPoints(geometry, webmercator.PToXY)
	}
}

// FromWebMercator takes a geometry encoded with WebMercator, and returns a Geometry encodes to the given srid.
func FromWebMercator(SRID uint64, geometry tegola.Geometry) (G, error) {
	switch SRID {
	default:
		return G{}, fmt.Errorf("Don't know how to convert from %v to %v.", SRID, tegola.WebMercator)
	case tegola.WebMercator:
		// Instead of just returning the geometry, we are cloning it so that the user of the API can rely
		// on the result to alway be a copy. Instead of being a reference in the on instance that it's already
		// in the same SRID.
		return CloneGeometry(geometry)
	case tegola.WGS84:
		return ApplyToPoints(geometry, webmercator.PToLonLat)
	}
}
