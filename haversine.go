package geohashquery

import blevegeo "github.com/blevesearch/bleve/v2/geo"

// haversine returns the great-circle distance in meters between two GPS
// coordinates using the Haversine formula.
func haversine(a, b GPSPoint) float64 {
	return blevegeo.Haversin(a.Lon, a.Lat, b.Lon, b.Lat) * 1000
}
