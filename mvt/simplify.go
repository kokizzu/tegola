package mvt

import (
	"context"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/planar"
	"github.com/go-spatial/geom/planar/simplify"
)

func SimplifyGeometryGeom(ctx context.Context, g geom.Geometry, tolerance float64) (geom.Geometry, error) {
	dp := simplify.DouglasPeucker{
		Tolerance: tolerance,
	}
	return planar.Simplify(ctx, dp, g)
}
