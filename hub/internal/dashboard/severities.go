package dashboard

// severityLabel upper-cases a severity for the compact ERR/WARN/INFO cell
// display, keeping unknown severities readable as-is.
func severityLabel(severity string) string {
	switch severity {
	case "error", "critical":
		return "ERR"
	case "warning":
		return "WARN"
	case "info":
		return "INFO"
	default:
		return severity
	}
}

// severityTone picks the Tailwind text tone per severity: errors read red,
// warnings amber, everything else muted.
func severityTone(severity string) string {
	switch severity {
	case "error", "critical":
		return "text-red-600 dark:text-red-400"
	case "warning":
		return "text-amber-600 dark:text-amber-400"
	default:
		return "text-sky-600 dark:text-sky-400"
	}
}
