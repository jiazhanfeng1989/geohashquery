package geohashquery

import (
	"errors"
	"math"

	"github.com/mmcloughlin/geohash"
	"github.com/twpayne/go-geos"
)

// GetQueryDirection determines the iteration direction based on the travel
// direction from startLocation to endLocation. The returned direction can be
// passed to query functions so that geohash cells are yielded in an order
// that matches the travel direction.
func GetQueryDirection(startLocation GPSPoint, endLocation GPSPoint) QueryDirection {
	fromWestToEast := !(endLocation.Lon > startLocation.Lon)
	fromSouthToNorth := !(endLocation.Lat > startLocation.Lat)

	if fromWestToEast {
		if fromSouthToNorth {
			return SouthWest
		}
		return NorthWest
	}
	if fromSouthToNorth {
		return SouthEast
	}
	return NorthEast
}

const geoHashSuffix = "bcfguvyz89destwx2367kmqr0145hjnp"

// EncodeGeoHashId encodes a GPS point into a GeoId at the given precision level.
func EncodeGeoHashId(point GPSPoint, geoHashLevel GeoHashLevelType) GeoId {
	return geohash.EncodeWithPrecision(point.Lat, point.Lon, uint(geoHashLevel))
}

type iterationOrder struct {
	latStart, latEnd, latStep float64
	lonStart, lonEnd, lonStep float64
}

func getIterationOrder(latSteps, lonSteps float64, direction QueryDirection) iterationOrder {
	switch direction {
	case SouthWest:
		return iterationOrder{0, latSteps, 1, 0, lonSteps, 1}
	case SouthEast:
		return iterationOrder{0, latSteps, 1, lonSteps, 0, -1}
	case NorthWest:
		return iterationOrder{latSteps, 0, -1, 0, lonSteps, 1}
	case NorthEast:
		return iterationOrder{latSteps, 0, -1, lonSteps, 0, -1}
	default:
		return iterationOrder{0, latSteps, 1, 0, lonSteps, 1}
	}
}

// GetGeoHashIdsByBoxWithDirection returns all geohash cells at the given
// precision level that fall inside the bounding box defined by
// (minLat, minLon) – (maxLat, maxLon). When originBuffer is non-nil, only
// cells that intersect the GEOS geometry are returned. queryDirection
// controls the iteration order of the returned cells.
func GetGeoHashIdsByBoxWithDirection(
	originBuffer *geos.Geom,
	minLat float64,
	minLon float64,
	maxLat float64,
	maxLon float64,
	geoHashLevel GeoHashLevelType,
	queryDirection QueryDirection,
) (GeoIds, error) {
	hashSouthWest := geohash.EncodeWithPrecision(minLat, minLon, uint(geoHashLevel))
	hashNorthEast := geohash.EncodeWithPrecision(maxLat, maxLon, uint(geoHashLevel))
	boxSourceWest := geohash.BoundingBox(hashSouthWest)
	boxNorthEast := geohash.BoundingBox(hashNorthEast)

	perLat := boxSourceWest.MaxLat - boxSourceWest.MinLat
	perLon := boxSourceWest.MaxLng - boxSourceWest.MinLng

	latStep := math.Round((boxNorthEast.MinLat - boxSourceWest.MinLat) / perLat)

	lonDiff := boxNorthEast.MinLng - boxSourceWest.MinLng
	if lonDiff < -180 {
		lonDiff += 360
	} else if lonDiff > 180 {
		lonDiff -= 360
	}
	lonStep := math.Round(lonDiff / perLon)

	if latStep < 0 {
		latStep = -latStep
	}
	if lonStep < 0 {
		lonStep = -lonStep
	}

	estimatedSize := int((latStep + 1) * (lonStep + 1))
	if estimatedSize < 0 || estimatedSize > 1000000 {
		estimatedSize = 1000
	}

	geoIds := make(GeoIds, 0, estimatedSize)
	seen := make(map[string]bool, estimatedSize)

	order := getIterationOrder(latStep, lonStep, queryDirection)
	centerLat := (boxSourceWest.MaxLat + boxSourceWest.MinLat) / 2
	centerLon := (boxSourceWest.MaxLng + boxSourceWest.MinLng) / 2

	for lat := order.latStart; ; lat += order.latStep {
		if order.latStep > 0 && lat > order.latEnd || order.latStep < 0 && lat < order.latEnd {
			break
		}

		for lon := order.lonStart; ; lon += order.lonStep {
			if order.lonStep > 0 && lon > order.lonEnd || order.lonStep < 0 && lon < order.lonEnd {
				break
			}

			neighborLat := ensureValidLat(centerLat + lat*perLat)
			neighborLon := ensureValidLon(centerLon + lon*perLon)
			neighborHash := geohash.EncodeWithPrecision(neighborLat, neighborLon, uint(geoHashLevel))

			if seen[neighborHash] {
				continue
			}
			seen[neighborHash] = true

			if originBuffer != nil {
				neighborBox := geohash.BoundingBox(neighborHash)
				if checkNeighborBoxIntersection(neighborBox, originBuffer) {
					geoIds = append(geoIds, neighborHash)
				}
			} else {
				geoIds = append(geoIds, neighborHash)
			}
		}
	}

	if len(geoIds) == 0 {
		return nil, errors.New("bounding box has no geoHash")
	}

	return geoIds, nil
}

func checkNeighborBoxIntersection(neighborBox geohash.Box, originBuffer *geos.Geom) bool {
	polygon := geos.NewPolygon([][][]float64{{
		{neighborBox.MinLng, neighborBox.MinLat},
		{neighborBox.MaxLng, neighborBox.MinLat},
		{neighborBox.MaxLng, neighborBox.MaxLat},
		{neighborBox.MinLng, neighborBox.MaxLat},
		{neighborBox.MinLng, neighborBox.MinLat},
	}})
	return polygon.Intersects(originBuffer)
}

// GetGeoHashIdsByBox is a convenience wrapper around
// GetGeoHashIdsByBoxWithDirection that defaults to SouthWest iteration order.
func GetGeoHashIdsByBox(
	originBuffer *geos.Geom,
	minLat float64,
	minLon float64,
	maxLat float64,
	maxLon float64,
	geoHashLevel GeoHashLevelType,
) (GeoIds, error) {
	return GetGeoHashIdsByBoxWithDirection(originBuffer, minLat, minLon, maxLat, maxLon, geoHashLevel, SouthWest)
}

// GetGeoHashIdsByRadius returns all geohash cells at the given precision
// level whose area intersects a circle centered at center with the given
// radius (in meters). The radius is clamped to [MinRadius, MaxRadius].
// queryDirection controls the iteration order of the returned cells.
func GetGeoHashIdsByRadius(center GPSPoint, radius int, geoHashLevel GeoHashLevelType, queryDirection QueryDirection) (GeoIds, error) {
	if radius > MaxRadius {
		radius = MaxRadius
	}

	if radius < MinRadius {
		radius = MinRadius
	}

	// Convert radius from meters to degrees (approximate).
	// At equator: 1 degree ~ 111.32 km.
	radiusInDegrees := float64(radius) / 111320.0

	// Adjust for latitude (longitude degrees get smaller away from equator).
	latRadians := center.Lat * math.Pi / 180.0
	cosLat := math.Cos(latRadians)
	if cosLat < 0.0001 {
		cosLat = 0.0001
	}
	lonRadiusInDegrees := radiusInDegrees / cosLat

	minLat := center.Lat - radiusInDegrees
	maxLat := center.Lat + radiusInDegrees
	minLon := center.Lon - lonRadiusInDegrees
	maxLon := center.Lon + lonRadiusInDegrees

	if minLat < MinLat {
		minLat = MinLat
	}
	if maxLat > MaxLat {
		maxLat = MaxLat
	}

	angle := float64(radius) / 10000 * 0.091
	point := geos.NewPointFromXY(center.Lon, center.Lat)
	buffer := point.Buffer(angle, 8)

	// Handle date line crossing.
	// GEOS works in planar coordinates and doesn't understand spherical
	// wrapping, so we split into two bounding-box queries and filter with
	// haversine distance instead.
	if minLon < MinLon || maxLon > MaxLon {
		var allGeoIds GeoIds
		seen := make(map[GeoId]bool)

		if minLon < MinLon {
			westMinLng := minLon + 360
			geoIds1, err := GetGeoHashIdsByBoxWithDirection(nil, minLat, westMinLng, maxLat, float64(MaxLon), geoHashLevel, queryDirection)
			if err == nil {
				for _, id := range geoIds1 {
					if !seen[id] && isWithinRadius(id, center, radius) {
						seen[id] = true
						allGeoIds = append(allGeoIds, id)
					}
				}
			}

			geoIds2, err := GetGeoHashIdsByBoxWithDirection(nil, minLat, float64(MinLon), maxLat, maxLon, geoHashLevel, queryDirection)
			if err == nil {
				for _, id := range geoIds2 {
					if !seen[id] && isWithinRadius(id, center, radius) {
						seen[id] = true
						allGeoIds = append(allGeoIds, id)
					}
				}
			}
		} else if maxLon > MaxLon {
			geoIds1, err := GetGeoHashIdsByBoxWithDirection(nil, minLat, minLon, maxLat, float64(MaxLon), geoHashLevel, queryDirection)
			if err == nil {
				for _, id := range geoIds1 {
					if !seen[id] && isWithinRadius(id, center, radius) {
						seen[id] = true
						allGeoIds = append(allGeoIds, id)
					}
				}
			}

			westMaxLng := maxLon - 360
			geoIds2, err := GetGeoHashIdsByBoxWithDirection(nil, minLat, float64(MinLon), maxLat, westMaxLng, geoHashLevel, queryDirection)
			if err == nil {
				for _, id := range geoIds2 {
					if !seen[id] && isWithinRadius(id, center, radius) {
						seen[id] = true
						allGeoIds = append(allGeoIds, id)
					}
				}
			}
		}

		if len(allGeoIds) == 0 {
			return nil, errors.New("bounding box has no geoHash")
		}
		return allGeoIds, nil
	}

	return GetGeoHashIdsByBoxWithDirection(buffer, minLat, minLon, maxLat, maxLon, geoHashLevel, queryDirection)
}

func isWithinRadius(geoId GeoId, center GPSPoint, radius int) bool {
	box := geohash.BoundingBox(geoId)
	corners := []GPSPoint{
		{Lat: box.MinLat, Lon: box.MinLng},
		{Lat: box.MinLat, Lon: box.MaxLng},
		{Lat: box.MaxLat, Lon: box.MinLng},
		{Lat: box.MaxLat, Lon: box.MaxLng},
	}

	for _, corner := range corners {
		if haversine(center, corner) <= float64(radius) {
			return true
		}
	}

	geoCenter := GPSPoint{
		Lat: (box.MinLat + box.MaxLat) / 2,
		Lon: (box.MinLng + box.MaxLng) / 2,
	}
	return haversine(center, geoCenter) <= float64(radius)
}

// GetGeoHashIdsByRoute returns all geohash cells at the given precision level
// that intersect a buffered corridor around the route polyline. routeWidth is
// the corridor half-width in meters. queryDirection controls the iteration
// order of the returned cells.
func GetGeoHashIdsByRoute(route GPSPoints, routeWidth int, geoHashLevel GeoHashLevelType, queryDirection QueryDirection) (GeoIds, error) {
	coords := make([][]float64, 0, len(route))
	for _, v := range route {
		coords = append(coords, []float64{v.Lon, v.Lat})
	}

	angle := float64(routeWidth) / 10000 * 0.091
	lineString := geos.NewLineString(coords)
	buffer := lineString.Buffer(angle, 8)
	envelope := buffer.Envelope()
	boundary := envelope.Boundary()
	box := boundary.CoordSeq().ToCoords()

	minLat := box[0][1]
	maxLat := box[2][1]
	minLon := box[0][0]
	maxLon := box[2][0]

	crossesDateLine := (maxLon - minLon) > 180

	if crossesDateLine || minLon < MinLon || maxLon > MaxLon {
		var allGeoIds GeoIds
		seen := make(map[GeoId]bool)

		if crossesDateLine && minLon >= MinLon && maxLon <= MaxLon {
			minRouteLon := 180.0
			maxRouteLon := -180.0
			for _, pt := range route {
				if pt.Lon > maxRouteLon {
					maxRouteLon = pt.Lon
				}
				if pt.Lon < minRouteLon {
					minRouteLon = pt.Lon
				}
			}

			if maxRouteLon > 0 {
				eastMinLng := minLon + 360 - angle
				if eastMinLng > 180 {
					eastMinLng = 180 - angle
				}
				if eastMinLng < maxRouteLon-angle {
					eastMinLng = maxRouteLon - angle
				}

				geoIds1, err := GetGeoHashIdsByBoxWithDirection(nil, minLat, eastMinLng, maxLat, float64(MaxLon), geoHashLevel, queryDirection)
				if err == nil {
					for _, id := range geoIds1 {
						if !seen[id] && isWithinRouteBuffer(id, route, routeWidth) {
							seen[id] = true
							allGeoIds = append(allGeoIds, id)
						}
					}
				}
			}

			if minRouteLon < 0 {
				westMaxLng := maxLon - 360 + angle
				if westMaxLng < -180 {
					westMaxLng = -180 + angle
				}
				if westMaxLng > minRouteLon+angle {
					westMaxLng = minRouteLon + angle
				}

				geoIds2, err := GetGeoHashIdsByBoxWithDirection(nil, minLat, float64(MinLon), maxLat, westMaxLng, geoHashLevel, queryDirection)
				if err == nil {
					for _, id := range geoIds2 {
						if !seen[id] && isWithinRouteBuffer(id, route, routeWidth) {
							seen[id] = true
							allGeoIds = append(allGeoIds, id)
						}
					}
				}
			}
		} else if minLon < MinLon {
			westMinLng := minLon + 360
			geoIds1, err := GetGeoHashIdsByBoxWithDirection(nil, minLat, westMinLng, maxLat, float64(MaxLon), geoHashLevel, queryDirection)
			if err == nil {
				for _, id := range geoIds1 {
					if !seen[id] && isWithinRouteBuffer(id, route, routeWidth) {
						seen[id] = true
						allGeoIds = append(allGeoIds, id)
					}
				}
			}

			geoIds2, err := GetGeoHashIdsByBoxWithDirection(nil, minLat, float64(MinLon), maxLat, maxLon, geoHashLevel, queryDirection)
			if err == nil {
				for _, id := range geoIds2 {
					if !seen[id] && isWithinRouteBuffer(id, route, routeWidth) {
						seen[id] = true
						allGeoIds = append(allGeoIds, id)
					}
				}
			}
		} else if maxLon > MaxLon {
			minRouteLon := 180.0
			maxRouteLon := -180.0
			for _, pt := range route {
				if pt.Lon > maxRouteLon {
					maxRouteLon = pt.Lon
				}
				if pt.Lon < minRouteLon {
					minRouteLon = pt.Lon
				}
			}

			if maxRouteLon > 0 {
				minPosLon := 180.0
				for _, pt := range route {
					if pt.Lon > 0 && pt.Lon < minPosLon {
						minPosLon = pt.Lon
					}
				}
				eastMinLng := minPosLon - angle
				if eastMinLng < 0 {
					eastMinLng = 0
				}

				geoIds1, err := GetGeoHashIdsByBoxWithDirection(nil, minLat, eastMinLng, maxLat, float64(MaxLon), geoHashLevel, queryDirection)
				if err == nil {
					for _, id := range geoIds1 {
						if !seen[id] && isWithinRouteBuffer(id, route, routeWidth) {
							seen[id] = true
							allGeoIds = append(allGeoIds, id)
						}
					}
				}
			}

			if minRouteLon < 0 {
				westMaxLng := maxLon - 360
				if westMaxLng < minRouteLon+angle {
					westMaxLng = minRouteLon + angle
				}
				if westMaxLng > maxLon {
					westMaxLng = maxLon
				}
				geoIds2, err := GetGeoHashIdsByBoxWithDirection(nil, minLat, float64(MinLon), maxLat, westMaxLng, geoHashLevel, queryDirection)
				if err == nil {
					for _, id := range geoIds2 {
						if !seen[id] && isWithinRouteBuffer(id, route, routeWidth) {
							seen[id] = true
							allGeoIds = append(allGeoIds, id)
						}
					}
				}
			}
		}

		if len(allGeoIds) == 0 {
			return nil, errors.New("bounding box has no geoHash")
		}
		return allGeoIds, nil
	}

	return GetGeoHashIdsByBoxWithDirection(buffer, minLat, minLon, maxLat, maxLon, geoHashLevel, queryDirection)
}

func isWithinRouteBuffer(geoId GeoId, route GPSPoints, routeWidth int) bool {
	box := geohash.BoundingBox(geoId)

	geoCenter := GPSPoint{
		Lat: (box.MinLat + box.MaxLat) / 2,
		Lon: (box.MinLng + box.MaxLng) / 2,
	}

	minDist := math.MaxFloat64
	for i := range len(route) - 1 {
		dist := pointToSegmentDistance(geoCenter, route[i], route[i+1])
		if dist < minDist {
			minDist = dist
		}
	}

	return minDist <= float64(routeWidth)
}

func pointToSegmentDistance(point, segStart, segEnd GPSPoint) float64 {
	if segStart.Lat == segEnd.Lat && segStart.Lon == segEnd.Lon {
		return haversine(point, segStart)
	}

	distToStart := haversine(point, segStart)
	distToEnd := haversine(point, segEnd)

	minDist := distToStart
	if distToEnd < minDist {
		minDist = distToEnd
	}

	lonDiff := segEnd.Lon - segStart.Lon
	crossesDateLine := false
	if lonDiff > 180 {
		lonDiff -= 360
		crossesDateLine = true
	} else if lonDiff < -180 {
		lonDiff += 360
		crossesDateLine = true
	}

	numSamples := 12
	for i := 1; i < numSamples; i++ {
		t := float64(i) / float64(numSamples)
		midLat := segStart.Lat + t*(segEnd.Lat-segStart.Lat)
		midLon := segStart.Lon + t*lonDiff

		if crossesDateLine {
			if midLon > MaxLon {
				midLon -= 360
			} else if midLon < MinLon {
				midLon += 360
			}
		}

		midPoint := GPSPoint{
			Lat: midLat,
			Lon: midLon,
		}
		dist := haversine(point, midPoint)
		if dist < minDist {
			minDist = dist
		}
	}

	return minDist
}

func ensureValidLon(lon float64) float64 {
	if lon > MaxLon {
		return MinLon + lon - MaxLon
	}

	if lon < MinLon {
		return MaxLon + lon - MaxLon
	}

	return lon
}

func ensureValidLat(lat float64) float64 {
	if lat > MaxLat {
		return MaxLat
	}

	if lat < MinLat {
		return MinLat
	}

	return lat
}

// SplitGeoHash splits a geohash into all 32 child geohashes (one level deeper).
func SplitGeoHash(geoId GeoId) GeoIds {
	newGeoIds := make(GeoIds, 32)

	for i, c := range geoHashSuffix {
		newGeoIds[i] = geoId + string(c)
	}

	return newGeoIds
}
