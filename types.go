package geohashquery

// GPSPoint represents a geographic coordinate with latitude and longitude in degrees.
type GPSPoint struct {
	Lat float64
	Lon float64
}

// GPSPoints is a slice of GPSPoint, typically representing a route or polyline.
type GPSPoints = []GPSPoint

// GeoHashLevelType represents the precision level (number of characters) of a geohash.
type GeoHashLevelType uint

const (
	GeoHashLevel3 = GeoHashLevelType(3)
	GeoHashLevel4 = GeoHashLevelType(4)
	GeoHashLevel5 = GeoHashLevelType(5)
)

// GeoId is a geohash string identifier.
type GeoId string

// GeoIds is a slice of GeoId.
type GeoIds []GeoId

// QueryDirection controls the iteration order when enumerating geohash cells
// inside a bounding box. Choosing the direction that matches the travel
// direction of a route lets callers receive geohash cells in a spatially
// coherent order (closest to the origin first).
type QueryDirection = int

const (
	NorthEast = QueryDirection(0)
	NorthWest = QueryDirection(1)
	SouthEast = QueryDirection(2)
	SouthWest = QueryDirection(3)
)

const (
	MinLat = -90
	MaxLat = 90
	MinLon = -180
	MaxLon = 180

	MaxRadius = 250000
	MinRadius = 10000
)
