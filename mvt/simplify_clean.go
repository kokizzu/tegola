package mvt

import (
	"context"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/tegola"
	"github.com/go-spatial/tegola/internal/convert"
	"github.com/go-spatial/tegola/maths/validate"
)

func CleanSimplifyGeometry(ctx context.Context, g tegola.Geometry, extent *geom.Extent, tolerance float64, simplify bool) (geo tegola.Geometry, err error) {
	if g == nil {
		return nil, nil
	}

	geomg, err := convert.ToGeom(g)
	if err != nil {
		return nil, err
	}

	if g == nil {
		return nil, nil
	}

	if simplify {
		geomg, err = SimplifyGeometryGeom(ctx, geomg, tolerance)
		if err != nil {
			return nil, err
		}
		if geomg == nil {
			return nil, nil
		}
	}

	geomg, err = validate.CleanGeometryGeom(ctx, geomg, extent)
	if err != nil {
		return nil, err
	}

	return convert.ToTegola(geomg)
}
