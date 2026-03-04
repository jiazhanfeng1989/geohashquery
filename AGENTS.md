# Agent Guide — geohashquery

## Project Overview

`geohashquery` is a standalone Go library for spatial querying of geohash cells. Given a geographic shape (circle, polyline corridor, or bounding box), it returns all geohash cells at a chosen precision level that intersect that shape.

- **Module path**: `github.com/jiazhanfeng1989/geohashquery`
- **Go version**: 1.24+
- **Package name**: `geohashquery`

## Tech Stack

| Category | Technology |
|----------|------------|
| Language | Go 1.24 |
| Geohash encoding | github.com/mmcloughlin/geohash |
| Geometry (CGo) | github.com/twpayne/go-geos (libgeos bindings) |
| Testing | stdlib testing |
| Lint | golangci-lint v2.5 (.golangci.yaml) |

## CGo Prerequisite

This library depends on `go-geos`, which requires **libgeos** system library. Users must install it before building:

- **macOS**: `brew install geos`
- **Ubuntu/Debian**: `sudo apt-get install libgeos-dev`
- **Verify**: `pkg-config --cflags --libs geos`

If `pkg-config` is not installed, install it first (`brew install pkg-config` or `sudo apt-get install pkg-config`).

## Directory Structure

```
geohashquery/
├── .golangci.yaml      # golangci-lint v2.5 configuration
├── go.mod              # Module definition
├── types.go            # GPSPoint, GeoId, GeoHashLevelType, QueryDirection, constants
├── haversine.go        # Haversine great-circle distance (internal)
├── geohash.go          # Core query functions + helpers
├── geohash_test.go     # Tests and benchmarks
├── README.md           # User-facing documentation
└── AGENTS.md           # This file
```

## Build, Test & Lint

```bash
# Build
go build ./...

# Test
go test ./...

# Test with verbose output
go test -v ./...

# Benchmarks
go test -bench=. ./...

# Lint
golangci-lint run ./...
```

## Post-Task Verification (IMPORTANT)

After completing any code change task, you **MUST** run the following two commands to verify correctness before considering the task done:

```bash
# 1. Lint check — all lint issues must be resolved
golangci-lint run ./...

# 2. Test verification — all tests must pass
go test ./...
```

If either command reports errors, fix the issues before finishing the task. Do not skip this step.

## API Quick Reference

### Primary Functions

| Function | Purpose |
|---|---|
| `GetGeoHashIdsByRadius(center, radius, level, direction)` | Geohash cells intersecting a circle (radius in meters, clamped to 10km–250km) |
| `GetGeoHashIdsByRoute(route, routeWidth, level, direction)` | Geohash cells intersecting a buffered polyline corridor (width in meters) |
| `GetGeoHashIdsByBoxWithDirection(geom, minLat, minLon, maxLat, maxLon, level, direction)` | Geohash cells inside a bounding box, optionally filtered by a GEOS geometry |

### Helper Functions

| Function | Purpose |
|---|---|
| `GetGeoHashIdsByBox(geom, minLat, minLon, maxLat, maxLon, level)` | Same as BoxWithDirection, defaults to SouthWest |
| `GetQueryDirection(start, end)` | Determine QueryDirection from two GPS points |
| `EncodeGeoHashId(point, level)` | Encode a GPS point to a GeoId |
| `SplitGeoHash(geoId)` | Split a geohash into 32 children (one level deeper) |

### Key Types

- `GPSPoint{Lat, Lon float64}` — geographic coordinate
- `GPSPoints = []GPSPoint` — polyline / route
- `GeoId string` — geohash string
- `GeoIds []GeoId` — slice of geohash strings
- `GeoHashLevelType uint` — precision level (3, 4, or 5)
- `QueryDirection int` — iteration order (`NorthEast`, `NorthWest`, `SouthEast`, `SouthWest`)

## Code Conventions

- Package name: `geohashquery` (single package, no sub-packages)
- Exported functions use PascalCase; internal helpers use camelCase
- All spatial coordinates use WGS 84 (latitude/longitude in degrees)
- The `haversine` function is unexported — it is an internal implementation detail
- Tests use the standard `testing` package only (no third-party assertion libraries)

## Important Notes

- All three primary functions correctly handle queries crossing the International Date Line (±180° longitude)
- Radius values are automatically clamped to `[MinRadius=10000, MaxRadius=250000]` meters
- The `QueryDirection` parameter affects only the **order** of returned geohash cells, not which cells are returned
- When `originBuffer` is nil in `GetGeoHashIdsByBoxWithDirection`, all cells in the bounding box are returned without geometry filtering
