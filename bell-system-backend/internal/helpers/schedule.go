package helpers

import (
	"sort"

	"arabiyya.edu.mv/bell-system-backend/internal/models"
)

// DayNames maps day numbers (1-7) to day names where 1=Sunday, 7=Saturday.
// Index 0 is intentionally empty to allow direct indexing by day number.
var DayNames = [8]string{"", "Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}

// MapDaysToInfo converts a slice of day numbers (1=Sunday, 7=Saturday) to
// a sorted slice of ScheduleDayInfo. Invalid day numbers outside 1-7 are
// silently filtered out.
func MapDaysToInfo(days []int) []models.ScheduleDayInfo {
	if days == nil {
		return nil
	}

	dayInfo := make([]models.ScheduleDayInfo, 0, len(days))
	for _, dayNum := range days {
		if dayNum >= 1 && dayNum <= 7 {
			dayInfo = append(dayInfo, models.ScheduleDayInfo{
				DayNumber: dayNum,
				DayName:   DayNames[dayNum],
			})
		}
	}

	sort.Slice(dayInfo, func(i, j int) bool {
		return dayInfo[i].DayNumber < dayInfo[j].DayNumber
	})

	return dayInfo
}
