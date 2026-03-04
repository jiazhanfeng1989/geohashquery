package geohashquery

import (
	"testing"

	"github.com/mmcloughlin/geohash"
)

func TestGetGeoHashIdsByRadius(t *testing.T) {
	center := GPSPoint{Lat: 39.00561, Lon: -114.21887}
	radius := 250000

	geoIds, err := GetGeoHashIdsByRadius(center, radius, GeoHashLevel5, SouthWest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(geoIds) == 0 {
		t.Fatal("expected non-empty result")
	}

	t.Logf("radius query returned %d geohash cells", len(geoIds))
}

func TestGetGeoHashIdsByRadiusCrossingDateLine(t *testing.T) {
	center := GPSPoint{Lat: 0.0, Lon: 179.0}
	radius := 150000

	geoIds, err := GetGeoHashIdsByRadius(center, radius, GeoHashLevel5, SouthWest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hasPositiveLon := false
	hasNegativeLon := false
	for _, geoId := range geoIds {
		box := geohash.BoundingBox(string(geoId))
		centerLon := (box.MinLng + box.MaxLng) / 2
		if centerLon > 0 && centerLon < 180 {
			hasPositiveLon = true
		}
		if centerLon < 0 && centerLon > -180 {
			hasNegativeLon = true
		}
	}

	if !hasPositiveLon || !hasNegativeLon {
		t.Error("expected geohashes on both sides of the date line")
	}
	t.Logf("date line radius query returned %d geohash cells", len(geoIds))
}

func TestGetGeoHashIdsByRoute(t *testing.T) {
	route := GPSPoints{
		{Lat: 37.7749, Lon: -122.4194},
		{Lat: 37.8044, Lon: -122.2712},
		{Lat: 37.8716, Lon: -122.2727},
		{Lat: 37.9577, Lon: -122.3477},
		{Lat: 38.0194, Lon: -122.1350},
	}
	routeWidth := 10000

	geoIds, err := GetGeoHashIdsByRoute(route, routeWidth, GeoHashLevel5, NorthEast)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(geoIds) == 0 {
		t.Fatal("expected non-empty result")
	}

	t.Logf("route query returned %d geohash cells", len(geoIds))
}

func TestGetGeoHashIdsByRouteCrossingDateLine(t *testing.T) {
	route := GPSPoints{
		{Lat: 0.0, Lon: 178.0},
		{Lat: 0.5, Lon: 179.0},
		{Lat: 1.0, Lon: 180.0},
		{Lat: 1.5, Lon: -179.0},
		{Lat: 2.0, Lon: -178.0},
	}
	routeWidth := 10000

	geoIds, err := GetGeoHashIdsByRoute(route, routeWidth, GeoHashLevel5, NorthWest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hasPositiveLon := false
	hasNegativeLon := false
	for _, geoId := range geoIds {
		box := geohash.BoundingBox(string(geoId))
		centerLon := (box.MinLng + box.MaxLng) / 2
		if centerLon > 0 && centerLon < 180 {
			hasPositiveLon = true
		}
		if centerLon < 0 && centerLon > -180 {
			hasNegativeLon = true
		}
	}

	if !hasPositiveLon || !hasNegativeLon {
		t.Error("expected geohashes on both sides of the date line")
	}
	t.Logf("date line route query returned %d geohash cells", len(geoIds))
}

func TestGetGeoHashIdsByBoxWithDirection(t *testing.T) {
	// EU bounding box
	minLat := 36.0282697679
	minLon := -10.428057909
	maxLat := 71.2148560447
	maxLon := 69.6460467676

	geoIds, err := GetGeoHashIdsByBoxWithDirection(nil, minLat, minLon, maxLat, maxLon, GeoHashLevel3, SouthWest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(geoIds) == 0 {
		t.Fatal("expected non-empty result")
	}

	t.Logf("box query returned %d geohash cells", len(geoIds))
}

func TestGetGeoHashIdsByBoxWithDirectionOrder(t *testing.T) {
	minLat := 37.0
	minLon := -123.0
	maxLat := 38.0
	maxLon := -122.0

	geoIdsSW, err := GetGeoHashIdsByBoxWithDirection(nil, minLat, minLon, maxLat, maxLon, GeoHashLevel5, SouthWest)
	if err != nil {
		t.Fatalf("SouthWest: unexpected error: %v", err)
	}

	geoIdsNE, err := GetGeoHashIdsByBoxWithDirection(nil, minLat, minLon, maxLat, maxLon, GeoHashLevel5, NorthEast)
	if err != nil {
		t.Fatalf("NorthEast: unexpected error: %v", err)
	}

	if len(geoIdsSW) != len(geoIdsNE) {
		t.Fatalf("expected same count, got SW=%d NE=%d", len(geoIdsSW), len(geoIdsNE))
	}

	if len(geoIdsSW) > 1 && geoIdsSW[0] == geoIdsNE[0] {
		t.Error("different directions should produce different ordering")
	}
}

func TestGetQueryDirection(t *testing.T) {
	tests := []struct {
		start    GPSPoint
		end      GPSPoint
		expected QueryDirection
	}{
		{GPSPoint{0, 0}, GPSPoint{1, 1}, NorthEast},
		{GPSPoint{0, 0}, GPSPoint{1, -1}, NorthWest},
		{GPSPoint{0, 0}, GPSPoint{-1, 1}, SouthEast},
		{GPSPoint{0, 0}, GPSPoint{-1, -1}, SouthWest},
	}

	for _, tc := range tests {
		got := GetQueryDirection(tc.start, tc.end)
		if got != tc.expected {
			t.Errorf("GetQueryDirection(%v, %v) = %d, want %d", tc.start, tc.end, got, tc.expected)
		}
	}
}

func TestEncodeGeoHashId(t *testing.T) {
	point := GPSPoint{Lat: 50.07868, Lon: 14.46089}
	geoId := EncodeGeoHashId(point, GeoHashLevel5)
	if geoId != "u2fkc" {
		t.Errorf("expected u2fkc, got %s", geoId)
	}
}

func TestSplitGeoHash(t *testing.T) {
	geoId := GeoId("u17")
	children := SplitGeoHash(geoId)
	if len(children) != 32 {
		t.Fatalf("expected 32 children, got %d", len(children))
	}
	for _, child := range children {
		if len(child) != 4 {
			t.Errorf("expected length 4, got %d for %s", len(child), child)
		}
	}
}

func TestHaversine(t *testing.T) {
	// New York to London: ~5570 km
	nyc := GPSPoint{Lat: 40.7128, Lon: -74.0060}
	london := GPSPoint{Lat: 51.5074, Lon: -0.1278}
	dist := haversine(nyc, london)
	if dist < 5500000 || dist > 5600000 {
		t.Errorf("NYC to London distance = %.0f m, expected ~5570000 m", dist)
	}

	// Same point should be 0
	dist = haversine(nyc, nyc)
	if dist != 0 {
		t.Errorf("same point distance = %f, expected 0", dist)
	}
}

func BenchmarkGetGeoHashIdsByRadius(b *testing.B) {
	center := GPSPoint{Lat: 39.00561, Lon: -114.21887}
	radius := 250000

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetGeoHashIdsByRadius(center, radius, GeoHashLevel5, SouthWest)
	}
}

func BenchmarkGetGeoHashIdsByRoute(b *testing.B) {
	route := GPSPoints{
		{Lat: 37.7749, Lon: -122.4194},
		{Lat: 37.8044, Lon: -122.2712},
		{Lat: 37.8716, Lon: -122.2727},
		{Lat: 37.9577, Lon: -122.3477},
		{Lat: 38.0194, Lon: -122.1350},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetGeoHashIdsByRoute(route, 10000, GeoHashLevel5, NorthEast)
	}
}
