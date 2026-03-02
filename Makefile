run:
	go run main.go listen

enable-sctl:
	go install
	go run . service install --binary-path ~/go/bin/hyprdocked

disable-sctl:
	go run . service uninstall

logs-sctl:
	go run . service logs -f
