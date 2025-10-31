package utils


// ValidateLocation - Check if within radius
func ValidateLocation(supplierLat, supplierLon, employeeLat, employeeLon float64, radius int) bool {
	distance := CalculateDistance(supplierLat, supplierLon, employeeLat, employeeLon)
	return distance <= float64(radius)
}