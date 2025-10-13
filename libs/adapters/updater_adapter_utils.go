package adapters

// closeChannels closing channels
func (l *UpdaterAdapter) closeChannels() {
	l.logger.Debug("close channels", "trace", "agent-os-instance.manager_adapter.closeChannels")
	close(l.chanClose)
	l.isClosed = true
	l.wg.Done()
}
