package mvt

import (
	"github.com/go-spatial/geom"
	"github.com/go-spatial/tegola"
)

func (c *cursor) ProjectPoints(pts ...[2]float64) ([][2]float64, error) {
	ppts := make([][2]float64, len(pts))
	for i := range pts {
		pt, err := c.tile.ToPixel(tegola.WebMercator, pts[i])
		if err != nil {
			return nil, err
		}
		ppts[i] = pt
	}
	return ppts, nil
}

func (c *cursor) ProjectPointGroups(plys ...[][][2]float64) ([][][][2]float64, error) {
	var err error
	mpply := make([][][][2]float64, len(plys))
	for ii, ply := range plys {
		pply := make([][][2]float64, len(ply))
		for i := range ply {
			pply[i], err = c.ProjectPoints(ply[i]...)
			if err != nil {
				return nil, err
			}
		}
		mpply[ii] = pply
	}
	return mpply, nil
}

func (c *cursor) ProjectGeometry(geo geom.Geometry) (geom.Geometry, error) {
	switch g := geo.(type) {
	case geom.Point:
		mgg, err := c.ProjectPoints([2]float64(g))
		if err != nil {
			return nil, err
		}
		return geom.Point(mgg[0]), nil

	case geom.MultiPoint:
		mgg, err := c.ProjectPoints([][2]float64(g)...)
		if err != nil {
			return nil, err
		}
		return geom.MultiPoint(mgg), nil

	case geom.LineString:
		mgg, err := c.ProjectPoints([][2]float64(g)...)
		if err != nil {
			return nil, err
		}
		return geom.LineString(mgg), nil

	case geom.MultiLineString:
		mmgg, err := c.ProjectPointGroups([][][2]float64(g))
		if err != nil {
			return nil, err
		}
		return geom.MultiLineString(mmgg[0]), nil

	case geom.Polygon:
		mmgg, err := c.ProjectPointGroups([][][2]float64(g))
		if err != nil {
			return nil, err
		}
		return geom.Polygon(mmgg[0]), nil

	case geom.MultiPolygon:
		mmgg, err := c.ProjectPointGroups([][][][2]float64(g)...)
		if err != nil {
			return nil, err
		}
		return geom.MultiPolygon(mmgg), nil
	default:
		return geo, ErrUnknownGeometryType{Geom: g}
	}
}
