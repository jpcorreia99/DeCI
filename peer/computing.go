package peer

// DataSharing describes functions to share data in a bittorrent-like system.
type Computing interface {
	Compute(executable []byte, executionArgs []string, fileExtension string,
		inputs []byte, numberOfRequestedNodes uint) ([]byte, error)
}
