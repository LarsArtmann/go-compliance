package cli

import (
	"os"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"

	"github.com/larsartmann/go-output"
)

// colorService styles table cells. One authority decides colored-vs-plain
// (this service) and the table renderer is told the same mode, so chrome
// and cells never disagree on a pipe whose TERM suggests color.
type colorService struct {
	enabled bool
}

// success/error/warning/info styles (fleet palette).
//
//nolint:gochecknoglobals // immutable fleet-palette styles, built once at init
var (
	styleSuccess = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981"))
	styleError   = lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444"))
	styleWarning = lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B"))
	styleInfo    = lipgloss.NewStyle().Foreground(lipgloss.Color("#38BDF8"))
	styleHeader  = lipgloss.NewStyle().Bold(true)
)

// newColorService decides color usage for the given mode, honoring
// NO_COLOR (https://no-color.org/) and TTY detection under auto.
func newColorService(mode output.ColorMode) *colorService {
	if os.Getenv("NO_COLOR") != "" {
		return &colorService{enabled: false}
	}

	switch mode {
	case output.ColorModeAlways:
		return &colorService{enabled: true}
	case output.ColorModeNever:
		return &colorService{enabled: false}
	case output.ColorModeAuto:
		return &colorService{
			enabled: colorprofile.Detect(os.Stdout, os.Environ()) != colorprofile.Ascii,
		}
	default:
		return &colorService{enabled: false}
	}
}

func (c *colorService) Success(s string) string { return c.apply(styleSuccess, s) }
func (c *colorService) Error(s string) string   { return c.apply(styleError, s) }
func (c *colorService) Warning(s string) string { return c.apply(styleWarning, s) }
func (c *colorService) Info(s string) string    { return c.apply(styleInfo, s) }
func (c *colorService) Header(s string) string  { return c.apply(styleHeader, s) }

func (c *colorService) apply(style lipgloss.Style, s string) string {
	if !c.enabled {
		return s
	}

	return style.Render(s)
}

// colorMode mirrors the service decision so table chrome and ANSI cells
// always agree: fully colored or fully plain.
func (c *colorService) colorMode() output.ColorMode {
	if c.enabled {
		return output.ColorModeAlways
	}

	return output.ColorModeNever
}
