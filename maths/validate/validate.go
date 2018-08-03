package validate

import (
	"context"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/tegola"
	"github.com/go-spatial/tegola/internal/convert"
)

func CleanGeometry(ctx context.Context, g tegola.Geometry, extent *geom.Extent) (geo tegola.Geometry, err error) {
	if g == nil {
		return nil, nil
	}

	geomg, err := convert.ToGeom(g)
	if err != nil {
		return nil, err
	}
	gg, err := CleanGeometryGeom(ctx, geomg, extent)
	if err != nil {
		return nil, err
	}
	return convert.ToTegola(gg)
}
