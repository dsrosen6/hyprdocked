package main

type set[E comparable] map[E]struct{}

func newSet[E comparable]() set[E] {
	return set[E]{}
}

func (s set[E]) contains(v E) bool {
	_, ok := s[v]
	return ok
}

func (s set[E]) add(vals ...E) {
	for _, v := range vals {
		s[v] = struct{}{}
	}
}

func strToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func lidStateToPtr(ls lidState) *lidState {
	if ls == lidStateUnknown {
		return nil
	}
	return &ls
}

func powerStateToPtr(ps powerState) *powerState {
	if ps == powerStateUnknown {
		return nil
	}
	return &ps
}
