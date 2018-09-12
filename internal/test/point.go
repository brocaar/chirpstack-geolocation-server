package test

import "math"

// core to sealevel distance
const coreToSeaLevel float64 = 6370000

// Point defines a geo point.
type Point struct {
	x float64
	y float64
	z float64
}

// NewPoint creates a new Point.
func NewPoint(lat, long, alt float64) Point {
	lat = degToRadian(lat)
	long = degToRadian(long)
	alt += coreToSeaLevel

	return Point{
		x: alt * math.Cos(lat) * math.Sin(long),
		y: alt * math.Sin(lat),
		z: alt * math.Cos(lat) * math.Cos(long),
	}
}

// Distance returns the distance between this and a given point.
// Note that this is the direct distance, without taking the curve of the earth into account!
func (p Point) Distance(other Point) float64 {
	return math.Sqrt(math.Pow(p.x-other.x, 2) + math.Pow(p.y-other.y, 2) + math.Pow(p.z-other.z, 2))
}

// LatLngAlt returns the latitude, longitude and altitude of the Point.
func (p Point) LatLngAlt() (float64, float64, float64) {
	r := math.Sqrt(math.Pow(p.x, 2) + math.Pow(p.y, 2) + math.Pow(p.z, 2))
	lat := radianToDeg(math.Asin(p.y / r))
	long := radianToDeg(math.Atan2(p.x, p.z))
	return lat, long, r - coreToSeaLevel
}

// degToRadian converts degrees into radians.
func degToRadian(deg float64) float64 {
	return deg * (math.Pi / 180.0)
}

// radianToDeg converts radians to degrees.
func radianToDeg(rad float64) float64 {
	return rad * (180.0 / math.Pi)
}
