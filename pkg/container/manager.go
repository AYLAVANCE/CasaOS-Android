package container

type Manager struct {
	primary  Runtime // Doki
	fallback Runtime // QEMU
	active   Runtime // Currently in use
}

func NewManager() *Manager {
	doki := NewDokiRuntime()
	qemu := NewQEMURuntime()

	m := &Manager{primary: doki, fallback: qemu}

	if doki.IsAvailable() {
		m.active = doki
	} else if qemu.IsAvailable() {
		m.active = qemu
	} else {
		// Default to doki if neither are truly available but user might install later
		m.active = doki
	}

	return m
}

func (m *Manager) GetActive() Runtime { return m.active }
func (m *Manager) SwitchToFallback()  { m.active = m.fallback }
