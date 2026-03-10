package handlers

// ScheduleHandler handles schedule item management HTTP requests.
type ScheduleHandler struct {
	Items ScheduleItemRepository
	Days  ScheduleDayRepository
}

// NewScheduleHandler creates a new ScheduleHandler.
func NewScheduleHandler(items ScheduleItemRepository, days ScheduleDayRepository) *ScheduleHandler {
	return &ScheduleHandler{Items: items, Days: days}
}
