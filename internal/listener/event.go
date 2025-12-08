package listener

type Event struct {
	Type    EventType
	Details string
}

type EventType string

const (
	DisplayAddEvent     EventType = "DISPLAY_ADDED"
	DisplayRemoveEvent  EventType = "DISPLAY_REMOVED"
	DisplayUnknownEvent EventType = "DISLAY_UNKNOWN_EVENT"
	ConfigUpdatedEvent  EventType = "CONFIG_UPDATED"
	LidSwitchOpenEvent  EventType = "LID_SWITCH_ON"
	LidSwitchCloseEvent EventType = "LID_SWITCH_CLOSED"
)
