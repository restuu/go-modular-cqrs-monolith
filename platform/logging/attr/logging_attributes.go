package attr

import "log/slog"

const (
	AttrKeyErr = "err"
)

func Err(err error) slog.Attr {
	return slog.Any(AttrKeyErr, err)
}
