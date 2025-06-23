package utils

import (
	"fmt"
)

func PrettyPrintTime(totalSeconds int) string {
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	formattedTime := ""
	if hours > 0 {
		formattedTime += fmt.Sprintf("%d hours, ", hours)
	}
	if minutes > 0 || hours > 0 {
		formattedTime += fmt.Sprintf("%d minutes, ", minutes)
	}
	formattedTime += fmt.Sprintf("%d seconds", seconds)

	return formattedTime
}
