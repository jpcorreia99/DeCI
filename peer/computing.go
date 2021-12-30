package peer

// DataSharing describes functions to share data in a bittorrent-like system.
type Computing interface {
	SampleFunc()

	Compute(executable []byte, data []byte) ([]byte, error)
}
