package main

// Monitor matches the output of 'hyprctl monitors', and is also used for config.
type Monitor struct {
	ID          int64   `json:"id,omitempty"`
	Name        string  `json:"name,omitempty"`
	Description string  `json:"description,omitempty"`
	Make        string  `json:"make,omitempty"`
	Model       string  `json:"model,omitempty"`
	Serial      string  `json:"serial,omitempty"`
	Width       int64   `json:"width,omitempty"`
	Height      int64   `json:"height,omitempty"`
	RefreshRate float64 `json:"refreshRate,omitempty"`
	X           int64   `json:"x,omitempty"`
	Y           int64   `json:"y,omitempty"`
	// ActiveWorkspace  Workspace `json:"activeWorkspace"`
	// SpecialWorkspace Workspace `json:"specialWorkspace"`
	Reserved        []int64  `json:"reserved,omitempty"`
	Scale           float64  `json:"scale,omitempty"`
	Transform       int64    `json:"transform,omitempty"`
	Focused         bool     `json:"focused,omitempty"`
	DPMSStatus      bool     `json:"dpmsStatus,omitempty"`
	Vrr             bool     `json:"vrr,omitempty"`
	Solitary        string   `json:"solitary,omitempty"`
	ActivelyTearing bool     `json:"activelyTearing,omitempty"`
	DirectScanoutTo string   `json:"directScanoutTo,omitempty"`
	Disabled        bool     `json:"disabled,omitempty"`
	CurrentFormat   string   `json:"currentFormat,omitempty"`
	MirrorOf        string   `json:"mirrorOf,omitempty"`
	AvailableModes  []string `json:"availableModes,omitempty"`
}

type Workspace struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}
