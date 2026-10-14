package platform

// PathProvider defines platform-specific directory paths
type PathProvider interface {
	ConfigDir() string   // Where casaos.conf lives
	DataDir() string     // Persistent data storage
	RuntimeDir() string  // PID files, sockets, URLs
	LogDir() string      // Log files
	DBDir() string       // SQLite database
	WWWDir() string      // Static web assets
	ShellDir() string    // Shell scripts/helpers
	StorageRoot() string // User data storage root
	UserDataDir() string // User configuration data
}

var Provider PathProvider
