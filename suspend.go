package main

var (
	suspendCmd = listenerEvent{Type: suspendCmdEvent}
	wakeCmd    = listenerEvent{Type: wakeCmdEvent}
)

func sendSuspendCmd() error {
	return sendCmd(string(suspendCmdEvent))
}

func sendWakeCmd() error {
	return sendCmd(string(wakeCmdEvent))
}
