copy-svc:
	cp ./hyprdocked-testing.service ~/.config/systemd/user/
	systemctl --user daemon-reload

update-svc:
	go install
	systemctl --user restart hyprdocked-testing.service
