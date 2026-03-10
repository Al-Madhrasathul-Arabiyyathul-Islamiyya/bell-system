package tests

import (
	"testing"

	"arabiyya.edu.mv/bell-system-backend/internal/helpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- MapDaysToInfo tests (spec: 1=Sunday through 7=Saturday) ---

func TestMapDaysToInfo_AllDays(t *testing.T) {
	days := []int{1, 2, 3, 4, 5, 6, 7}
	result := helpers.MapDaysToInfo(days)

	require.Len(t, result, 7)
	expected := []struct {
		num  int
		name string
	}{
		{1, "Sunday"},
		{2, "Monday"},
		{3, "Tuesday"},
		{4, "Wednesday"},
		{5, "Thursday"},
		{6, "Friday"},
		{7, "Saturday"},
	}
	for i, exp := range expected {
		assert.Equal(t, exp.num, result[i].DayNumber)
		assert.Equal(t, exp.name, result[i].DayName)
	}
}

func TestMapDaysToInfo_Nil(t *testing.T) {
	result := helpers.MapDaysToInfo(nil)
	assert.Nil(t, result)
}

func TestMapDaysToInfo_Empty(t *testing.T) {
	result := helpers.MapDaysToInfo([]int{})
	assert.Empty(t, result)
}

func TestMapDaysToInfo_SortedOutput(t *testing.T) {
	days := []int{5, 2, 7, 1}
	result := helpers.MapDaysToInfo(days)

	require.Len(t, result, 4)
	assert.Equal(t, 1, result[0].DayNumber) // Sunday
	assert.Equal(t, 2, result[1].DayNumber) // Monday
	assert.Equal(t, 5, result[2].DayNumber) // Thursday
	assert.Equal(t, 7, result[3].DayNumber) // Saturday
}

func TestMapDaysToInfo_InvalidDaysFiltered(t *testing.T) {
	days := []int{0, 1, 8, 3, -1}
	result := helpers.MapDaysToInfo(days)

	require.Len(t, result, 2)
	assert.Equal(t, 1, result[0].DayNumber)
	assert.Equal(t, 3, result[1].DayNumber)
}

func TestMapDaysToInfo_WeekdaysOnly(t *testing.T) {
	days := []int{2, 3, 4, 5, 6}
	result := helpers.MapDaysToInfo(days)

	require.Len(t, result, 5)
	assert.Equal(t, "Monday", result[0].DayName)
	assert.Equal(t, "Friday", result[4].DayName)
}

func TestMapDaysToInfo_Duplicates(t *testing.T) {
	days := []int{2, 2, 3}
	result := helpers.MapDaysToInfo(days)
	assert.Len(t, result, 3)
}

func TestDayNames_Array(t *testing.T) {
	assert.Equal(t, "", helpers.DayNames[0])
	assert.Equal(t, "Sunday", helpers.DayNames[1])
	assert.Equal(t, "Saturday", helpers.DayNames[7])
}
