package utils

func BuildClientKey(passengerID string, driverID string) string {
	if driverID != "" {
		return driverID
	}

	return passengerID
}
