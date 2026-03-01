copy-svc:
	cp ./hyprdocked-testing.service ~/.config/systemd/user/
	systemctl --user daemon-reload
