package base

const SGLogContextKey = "sg_log_context"

// LogContext stores values which may be useful to include in logs
type LogContext struct {
	// CorrelationID is a pre-formatted identifier used to correlate logs.
	// E.g: Either blip context ID or HTTP Serial number.
	CorrelationID           string
}

// addContext returns a string format with additional log context if present.
func (lc *LogContext) addContext(format string) string {
	if lc == nil {
		return ""
	}

	if lc.CorrelationID != "" {
		format = "cID=" + lc.CorrelationID + " " + format
	}

	return format
}
