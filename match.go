package main

// matchesIdentifiers verifies if a user-provided set of monitor identifiers has a match
// for a monitor fetched from Hyprland. It only checks non-empty fields in the identifier set,
// since the user may provide only name, only description, or some other combination.
func matchesIdentifiers(hm monitor, ident monitorIdentifiers) bool {
	if ident.Name == "" && ident.Description == "" {
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

	return true
}
