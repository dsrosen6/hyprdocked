package main

func (h *hyprctlClient) listMonitors() (map[string]Monitor, error) {
	var monitors []Monitor
	if err := h.runCommandWithUnmarshal([]string{"monitors"}, &monitors); err != nil {
		return nil, err
	}

	mm := make(map[string]Monitor, len(monitors))
	for _, m := range monitors {
		mm[m.Name] = m
	}

	return mm, nil
}
