package main

// status is the combined status of the device's docked (external or just laptop),
// power (ac or battery) and lid (closed or opened).
type status int

const (
	statusUnknown status = iota
	statusOnlyLaptopClosed
	statusOnlyLaptopOpened
	statusDockedClosed
	statusDockedOpened
)

func (a *app) getStatus() status {
	return getStatus(a.laptopDisplay, a.allDisplays, a.state)
}

func (s status) string() string {
	switch s {
	case statusUnknown:
		return "unknown"
	case statusOnlyLaptopClosed:
		return "only_laptop_lid_closed"
	case statusOnlyLaptopOpened:
		return "only_laptop_lid_opened"
	case statusDockedClosed:
		return "docked_lid_closed"
	case statusDockedOpened:
		return "docked_lid_opened"
	default:
		return "unknown"
	}
}

func getStatus(laptopDisplay display, allDisplays []display, state *state) status {
	if displayReady(laptopDisplay) && len(allDisplays) <= 1 {
		return laptopOnlyStatus(state.lidState)
	}

	return dockedStatus(state.lidState)
}

func laptopOnlyStatus(ls lidState) status {
	switch ls {
	case lidStateClosed:
		return statusOnlyLaptopClosed
	case lidStateOpened:
		return statusOnlyLaptopOpened
	default:
		return statusUnknown
	}
}

func dockedStatus(ls lidState) status {
	switch ls {
	case lidStateClosed:
		return statusDockedClosed
	case lidStateOpened:
		return statusDockedOpened
	default:
		return statusUnknown
	}
}
