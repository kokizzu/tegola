package validate

import (
	"context"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/planar/clip"
	"github.com/go-spatial/geom/planar/makevalid"
	"github.com/go-spatial/geom/planar/makevalid/hitmap"
)

func CleanGeometryGeom(ctx context.Context, g geom.Geometry, extent *geom.Extent) (geo geom.Geometry, err error) {
	if g == nil {
		return nil, nil
	}

	hm, err := hitmap.New(nil, g)
	if err != nil {
		return nil, err
	}

	mkv := makevalid.Makevalid{
		Hitmap:  hm,
		Clipper: clip.Default,
	}
	geo, _, err = mkv.Makevalid(ctx, g, extent)

	return geo, err
}
