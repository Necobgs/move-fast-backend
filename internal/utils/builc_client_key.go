package utils

func BuildClientKey(userId, driverId string) string {
	if driverId != "" {
		return "driver:" + driverId
	}
	return "user:" + userId
}
