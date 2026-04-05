package actions

import (
	"github.com/israelcorrea/crit-ide/internal/editor"
	"github.com/israelcorrea/crit-ide/internal/events"
)

// defaultContextMenuItems returns the standard context menu items.
func defaultContextMenuItems() []editor.MenuItem {
	return []editor.MenuItem{
		{Label: "Copy", ActionID: "clipboard.copy"},
		{Label: "Cut", ActionID: "clipboard.cut"},
		{Label: "Paste", ActionID: "clipboard.paste"},
		{IsSeparator: true},
		{Label: "Select All", ActionID: "select.all"},
	}
}

// --- Menu open action ---

type menuOpen struct{}

func (a *menuOpen) ID() string { return "menu.open" }

func (a *menuOpen) Run(ctx *ActionContext) error {
	payload, ok := ctx.Event.Payload.(events.MouseClickPayload)
	if !ok {
		return nil
	}
	items := defaultContextMenuItems()
	// Start with first non-separator item selected.
	selectedIdx := 0
	for i, item := range items {
		if !item.IsSeparator {
			selectedIdx = i
			break
		}
	}
	ctx.App.SetContextMenu(&editor.MenuState{
		ScreenX:     payload.ScreenX,
		ScreenY:     payload.ScreenY,
		Items:       items,
		SelectedIdx: selectedIdx,
	})
	ctx.App.SetInputMode(ModeContextMenu)
	return nil
}

// --- Menu close action ---

type menuClose struct{}

func (a *menuClose) ID() string { return "menu.close" }

func (a *menuClose) Run(ctx *ActionContext) error {
	ctx.App.SetContextMenu(nil)
	ctx.App.SetInputMode(ModeNormal)
	return nil
}

// --- Menu up action ---

type menuUp struct{}

func (a *menuUp) ID() string { return "menu.up" }

func (a *menuUp) Run(ctx *ActionContext) error {
	menu := ctx.App.ContextMenu()
	if menu == nil || len(menu.Items) == 0 {
		return nil
	}
	idx := menu.SelectedIdx
	for i := 0; i < len(menu.Items); i++ {
		idx--
		if idx < 0 {
			idx = len(menu.Items) - 1
		}
		if !menu.Items[idx].IsSeparator {
			menu.SelectedIdx = idx
			return nil
		}
	}
	return nil
}

// --- Menu down action ---

type menuDown struct{}

func (a *menuDown) ID() string { return "menu.down" }

func (a *menuDown) Run(ctx *ActionContext) error {
	menu := ctx.App.ContextMenu()
	if menu == nil || len(menu.Items) == 0 {
		return nil
	}
	idx := menu.SelectedIdx
	for i := 0; i < len(menu.Items); i++ {
		idx++
		if idx >= len(menu.Items) {
			idx = 0
		}
		if !menu.Items[idx].IsSeparator {
			menu.SelectedIdx = idx
			return nil
		}
	}
	return nil
}

// --- Menu execute action ---

type menuExecute struct{}

func (a *menuExecute) ID() string { return "menu.execute" }

func (a *menuExecute) Run(ctx *ActionContext) error {
	menu := ctx.App.ContextMenu()
	if menu == nil || len(menu.Items) == 0 {
		return nil
	}
	item := menu.Items[menu.SelectedIdx]
	if item.IsSeparator || item.ActionID == "" {
		return nil
	}
	// Close menu, then post the selected action for the main loop to execute.
	actionID := item.ActionID
	ctx.App.SetContextMenu(nil)
	ctx.App.SetInputMode(ModeNormal)
	ctx.App.PostAction(actionID)
	return nil
}

// --- Menu click action ---

// menuClick handles a mouse click while the context menu is open.
// If the click is inside the menu, it executes the clicked item.
// If outside, it closes the menu.
type menuClick struct{}

func (a *menuClick) ID() string { return "menu.click" }

func (a *menuClick) Run(ctx *ActionContext) error {
	payload, ok := ctx.Event.Payload.(events.MouseClickPayload)
	if !ok {
		return nil
	}
	menu := ctx.App.ContextMenu()
	if menu == nil {
		return nil
	}

	// Compute popup bounds (matching renderPopup logic).
	maxLabel := 0
	for _, item := range menu.Items {
		if !item.IsSeparator && len(item.Label) > maxLabel {
			maxLabel = len(item.Label)
		}
	}
	popupWidth := maxLabel + 4
	if popupWidth < 12 {
		popupWidth = 12
	}
	popupHeight := len(menu.Items) + 2

	screenW := ctx.App.ScreenWidth()
	vpHeight := ctx.App.ViewportHeight() + 1 // total screen height

	px := menu.ScreenX
	py := menu.ScreenY
	if px+popupWidth > screenW {
		px = screenW - popupWidth
	}
	if px < 0 {
		px = 0
	}
	if py+popupHeight > vpHeight {
		py = vpHeight - popupHeight
	}
	if py < 0 {
		py = 0
	}

	cx, cy := payload.ScreenX, payload.ScreenY

	// Check if click is inside the popup content area (excluding borders).
	if cx > px && cx < px+popupWidth-1 && cy > py && cy < py+popupHeight-1 {
		// Click is inside — determine which item.
		itemIdx := cy - py - 1 // -1 for top border.
		if itemIdx >= 0 && itemIdx < len(menu.Items) {
			item := menu.Items[itemIdx]
			if !item.IsSeparator && item.ActionID != "" {
				ctx.App.SetContextMenu(nil)
				ctx.App.SetInputMode(ModeNormal)
				ctx.App.PostAction(item.ActionID)
				return nil
			}
		}
		return nil // Click on separator or invalid — do nothing.
	}

	// Click is outside — close menu.
	ctx.App.SetContextMenu(nil)
	ctx.App.SetInputMode(ModeNormal)
	return nil
}
