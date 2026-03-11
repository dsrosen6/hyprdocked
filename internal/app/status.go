package app

import (
	"github.com/dsrosen6/hyprdocked/internal/hypr"
	"github.com/dsrosen6/hyprdocked/internal/power"
)

// status is the combined status of the device's docked state (external or just laptop),
// and lid (closed or opened).
type status int

const (
	statusUnknown status = iota
	statusOnlyLaptopClosed
	statusOnlyLaptopOpened
	statusDockedClosed
	statusDockedOpened
)

func (a *App) status() status {
	return getStatus(a.laptopDisplay, a.allDisplays, a.state)
}

func (a *App) statusString() string {
	return a.status().string()
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

func getStatus(laptopDisplay hypr.Monitor, allDisplays []hypr.Monitor, state *state) status {
	laptopEnabled := false
	for _, d := range allDisplays {
		if d.Name == laptopDisplay.Name {
			laptopEnabled = true
			break
		}
	}

	if displayReady(laptopDisplay) && (len(allDisplays) == 0 || (len(allDisplays) == 1 && laptopEnabled)) {
		return laptopOnlyStatus(state.lidState)
	}

	return dockedStatus(state.lidState)
}

func laptopOnlyStatus(ls power.LidState) status {
	switch ls {
	case power.LidStateClosed:
		return statusOnlyLaptopClosed
	case power.LidStateOpened:
		return statusOnlyLaptopOpened
	default:
		return statusUnknown
	}
}

func dockedStatus(ls power.LidState) status {
	switch ls {
	case power.LidStateClosed:
		return statusDockedClosed
	case power.LidStateOpened:
		return statusDockedOpened
	default:
		return statusUnknown
	}
}
