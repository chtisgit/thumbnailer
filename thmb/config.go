package thmb

// Config is the configuration data for the Server.
type Config struct {
	// Network is the network to listen on, as in the arguments
	// of net.Listen.
	Network string

	// Addr is the address for the server to listen on.
	Addr string

	// MaxSize is the maximum size in bytes for an image
	// that can be accepted for processing.
	MaxSize uint32

	// NumWorkers is the number of go routines that are spawned
	// to process the requests.
	NumWorkers int

	// TempDir is a temporary directory to use for uploads.
	TempDir string
}
