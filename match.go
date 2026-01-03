package main

type labeledMonitor struct {
	Label   string
	Monitor monitor
}

type labelLookup struct {
	monitors []labeledMonitor
	confirm  map[string]bool
}

func newLabelLookup(cfgMtrs monitorConfigMap, hyprMtrs []monitor) labelLookup {
	lm := matchMonitorsToLabels(cfgMtrs, hyprMtrs)
	confirm := make(map[string]bool, len(lm))
	for _, l := range lm {
		confirm[l.Label] = true
	}

	return labelLookup{
		monitors: lm,
		confirm:  confirm,
	}
}

func matchMonitorsToLabels(cfgMtrs monitorConfigMap, hyprMtrs []monitor) []labeledMonitor {
	var labeled []labeledMonitor
	for label, cfg := range cfgMtrs {
		for _, hm := range hyprMtrs {
			if matchesIdentifiers(hm, cfg.Identifiers) {
				labeled = append(labeled, labeledMonitor{
					Label:   label,
					Monitor: hm,
				})
				break
			}
		}
	}

	return labeled
}

func matchesIdentifiers(hm monitor, ident monitorIdentifiers) bool {
	if ident.Name == "" && ident.Description == "" &&
		ident.Make == "" && ident.Model == "" {
		return false
	}

	if ident.Name != "" {
		if ident.Name != hm.Name {
			return false
		}
	}

	if ident.Description != "" {
		if ident.Description != hm.Description {
			return false
		}
	}

	if ident.Make != "" {
		if ident.Make != hm.Make {
			return false
		}
	}

	if ident.Model != "" {
		if ident.Model != hm.Model {
			return false
		}
	}

	return true
}
