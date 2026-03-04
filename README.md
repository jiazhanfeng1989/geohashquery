# geohashquery

A Go library for spatial querying of [geohash](https://en.wikipedia.org/wiki/Geohash) cells. Given a geographic shape — circle, polyline corridor, or bounding box — the library returns all geohash cells at a chosen precision level that intersect that shape.

Typical use cases include:
- Finding candidate geohash cells for a radius search around a point.
- Discovering geohash cells along a driving route (with a configurable buffer width).
- Enumerating geohash cells inside an arbitrary bounding box.

## Prerequisites

This library uses [go-geos](https://github.com/twpayne/go-geos), which is a CGo binding to [libgeos](https://libgeos.org). You must install **libgeos** on your system before using this library.

### macOS

```bash
brew install geos
```

### Ubuntu / Debian

```bash
sudo apt-get install libgeos-dev
```

### Verify installation

```bash
pkg-config --cflags --libs geos
```

If `pkg-config` reports the GEOS flags correctly, you are ready to go.

## Install

```bash
go get github.com/jiazhanfeng1989/geohashquery
```

## API

### Core Query Functions

#### `GetGeoHashIdsByRadius`

```go
func GetGeoHashIdsByRadius(
    center GPSPoint,
    radius int,
    geoHashLevel GeoHashLevelType,
    queryDirection QueryDirection,
) (GeoIds, error)
```

Returns all geohash cells that intersect a circle centered at `center` with the given `radius` (in meters). The radius is clamped to `[10 000, 250 000]` meters. Correctly handles queries that cross the International Date Line.

#### `GetGeoHashIdsByRoute`

```go
func GetGeoHashIdsByRoute(
    route GPSPoints,
    routeWidth int,
    geoHashLevel GeoHashLevelType,
    queryDirection QueryDirection,
) (GeoIds, error)
```

Returns all geohash cells that intersect a buffered corridor around a polyline route. `routeWidth` specifies the corridor half-width in meters. Correctly handles routes that cross the International Date Line.

#### `GetGeoHashIdsByBoxWithDirection`

```go
func GetGeoHashIdsByBoxWithDirection(
    originBuffer *geos.Geom,
    minLat, minLon, maxLat, maxLon float64,
    geoHashLevel GeoHashLevelType,
    queryDirection QueryDirection,
) (GeoIds, error)
```

Returns all geohash cells that fall inside the given bounding box. When `originBuffer` is non-nil, only cells that intersect that GEOS geometry are included. `queryDirection` controls the iteration order — choose a direction that matches the travel direction so cells are returned in a spatially coherent order.

### Helper Functions

| Function | Description |
|---|---|
| `GetGeoHashIdsByBox` | Convenience wrapper around `GetGeoHashIdsByBoxWithDirection` with `SouthWest` direction. |
| `GetQueryDirection` | Determines `QueryDirection` from a start and end point. |
| `EncodeGeoHashId` | Encodes a GPS point into a `GeoId` at the given precision. |
| `SplitGeoHash` | Splits a geohash into all 32 child geohashes (one level deeper). |

### Types

```go
type GPSPoint struct {
    Lat float64
    Lon float64
}

type GPSPoints = []GPSPoint

type GeoId        string
type GeoIds       []GeoId
type GeoHashLevelType uint

// Predefined precision levels
const (
    GeoHashLevel3 = GeoHashLevelType(3) // ~156 km × 156 km cells
    GeoHashLevel4 = GeoHashLevelType(4) // ~39 km × 19.5 km cells
    GeoHashLevel5 = GeoHashLevelType(5) // ~4.9 km × 4.9 km cells
)

// Query directions
const (
    NorthEast = QueryDirection(0)
    NorthWest = QueryDirection(1)
    SouthEast = QueryDirection(2)
    SouthWest = QueryDirection(3)
)
```

## Examples

### Find geohash cells within a radius

```go
package main

import (
    "fmt"
    gq "github.com/jiazhanfeng1989/geohashquery"
)

func main() {
    center := gq.GPSPoint{Lat: 37.7749, Lon: -122.4194} // San Francisco
    radius := 50000 // 50 km

    geoIds, err := gq.GetGeoHashIdsByRadius(center, radius, gq.GeoHashLevel5, gq.SouthWest)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Found %d geohash cells within %d m of SF\n", len(geoIds), radius)
}
```

### Find geohash cells along a route

```go
package main

import (
    "fmt"
    gq "github.com/jiazhanfeng1989/geohashquery"
)

func main() {
    route := gq.GPSPoints{
        {Lat: 37.7749, Lon: -122.4194}, // San Francisco
        {Lat: 37.8044, Lon: -122.2712}, // Oakland
        {Lat: 38.5816, Lon: -121.4944}, // Sacramento
    }
    routeWidth := 10000 // 10 km buffer

    direction := gq.GetQueryDirection(route[0], route[len(route)-1])
    geoIds, err := gq.GetGeoHashIdsByRoute(route, routeWidth, gq.GeoHashLevel5, direction)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Found %d geohash cells along route\n", len(geoIds))
}
```

### Find geohash cells in a bounding box

```go
package main

import (
    "fmt"
    gq "github.com/jiazhanfeng1989/geohashquery"
)

func main() {
    // Europe bounding box
    geoIds, err := gq.GetGeoHashIdsByBoxWithDirection(
        nil,
        36.0, -10.5, 71.2, 69.6,
        gq.GeoHashLevel3,
        gq.SouthWest,
    )
    if err != nil {
        panic(err)
    }
    fmt.Printf("Found %d geohash cells in Europe\n", len(geoIds))
}
```

## License

MIT
