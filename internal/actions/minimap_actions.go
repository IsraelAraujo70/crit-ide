package actions

import (
	"github.com/israelcorrea/crit-ide/internal/events"
	"github.com/israelcorrea/crit-ide/internal/render"
)

// --- Minimap toggle action ---

type minimapToggle struct{}

func (a *minimapToggle) ID() string { return "minimap.toggle" }

func (a *minimapToggle) Run(ctx *ActionContext) error {
	ctx.App.ToggleMinimap()
	return nil
}

// --- Minimap click action ---

type minimapClick struct{}

func (a *minimapClick) ID() string { return "minimap.click" }

func (a *minimapClick) Run(ctx *ActionContext) error {
	payload, ok := ctx.Event.Payload.(events.MouseClickPayload)
	if !ok {
		return nil
	}

	buf := ctx.App.ActiveBuffer()
	scrollY := ctx.App.ScrollY()
	vpHeight := ctx.App.ViewportHeight()

	// contentStartY = tabBarHeight(1) + border(1) = 2
	contentStartY := 2
	lineCount := buf.Text.LineCount()

	targetLine := render.MinimapClickToLine(
		payload.ScreenY,
		contentStartY,
		vpHeight,
		lineCount,
		scrollY,
	)

	// Center the viewport on the target line.
	newScrollY := targetLine - vpHeight/2
	if newScrollY < 0 {
		newScrollY = 0
	}
	maxScroll := lineCount - vpHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	if newScrollY > maxScroll {
		newScrollY = maxScroll
	}
	ctx.App.SetScrollY(newScrollY)

	// Move cursor to target line.
	buf.SetCursorPos(targetLine, 0)

	return nil
}
