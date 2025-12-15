package enapterapi

import "io"

func DrainAndClose(rc io.ReadCloser) error {
	_, _ = io.Copy(io.Discard, rc)
	return rc.Close()
}
