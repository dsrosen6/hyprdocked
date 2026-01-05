package main

type status int

const (
	statusUnknown status = iota
	statusOnlyLaptopClosed
	statusOnlyLaptopOpened
	statusDockedClosed
	statusDockedOpened
)

func (a *app) getStatus() status {
	em := a.externalMonitors()
	return getStatus(em, a.currentState)
}

func (a *app) externalMonitors() []monitor {
	var em []monitor
	for _, m := range a.currentState.monitors {
		if matchesIdentifiers(m, a.cfg.Laptop.monitorIdentifiers) {
			continue
		}
		em = append(em, m)
	}

	return em
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

func getStatus(externals []monitor, state *state) status {
	if len(externals) == 0 {
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
