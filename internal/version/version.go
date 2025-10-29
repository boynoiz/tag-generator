package version

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const expectedVersionParts = 5

// Calculate determines the next version based on current tag and type
func Calculate(currentTag string, isFix bool, isStaging bool) (string, error) {
	// If no current tag exists, generate first version
	if currentTag == "" {
		newVersion := generateNew()
		return newVersion, nil
	}

	versionParts, err := split(currentTag)
	if err != nil || len(versionParts) < expectedVersionParts {
		newVersion := generateNew()
		return newVersion, nil
	}

	year := versionParts[0]
	month := versionParts[1]
	week := versionParts[2]
	release := versionParts[3]
	fix := versionParts[4]

	if isFix {
		fix = fix + 1
	} else {
		// Release version
		year, month, week, release, fix = calculateRelease(year, month, week, release)
	}

	return format(year, month, week, release, fix, isStaging), nil
}

func calculateRelease(currentYear, currentMonth, currentWeek, currentRelease int) (int, int, int, int, int) {
	year := currentYear
	month := currentMonth
	week := currentWeek
	release := currentRelease
	fix := 0

	if getCurrentYear() > year {
		year = getCurrentYear()
	}
	if getCurrentMonth() > month {
		month = getCurrentMonth()
	}
	if week != getWeekOfTheMonth() {
		week = getWeekOfTheMonth()
		release = 1
	} else {
		release = release + 1
	}

	return year, month, week, release, fix
}

func format(year, month, week, release, fix int, isStaging bool) string {
	version := fmt.Sprintf("%d.%d.%d.%d.%d", year, month, week, release, fix)
	if isStaging {
		return version + "-staging"
	}
	return version
}

func generateNew() string {
	return fmt.Sprintf("%d.%d.%d.%d.%d", getCurrentYear(), getCurrentMonth(), getWeekOfTheMonth(), 1, 0)
}

func split(version string) ([]int, error) {
	version = strings.TrimSpace(version)
	// Remove -staging suffix if present
	version = strings.TrimSuffix(version, "-staging")

	resultString := strings.Split(version, ".")
	result := make([]int, len(resultString))
	for idx, i := range resultString {
		convert, err := strconv.Atoi(strings.TrimSpace(i))
		if err != nil {
			return nil, fmt.Errorf("invalid version format: %w", err)
		}
		result[idx] = convert
	}
	return result, nil
}

func getCurrentYear() int {
	return time.Now().Year()
}

func getCurrentMonth() int {
	return int(time.Now().Month())
}

func getWeekOfTheMonth() int {
	now := time.Now()
	// Calculate week of month: (day of month - 1) / 7 + 1
	return (now.Day()-1)/7 + 1
}
