package display

import (
	runewidth "github.com/mattn/go-runewidth"
	"github.com/micro-editor/tcell/v2"
	"github.com/zyedidia/micro/v2/internal/config"
	"github.com/zyedidia/micro/v2/internal/screen"
	"github.com/zyedidia/micro/v2/internal/util"
)

// MenuItem represents a single menu item
type MenuItem struct {
	Name    string
	Action  string
	Hotkey  rune
	Enabled bool
}

// MenuWindow displays a horizontal menu bar at the top of the screen
type MenuWindow struct {
	MenuItems     []MenuItem
	Active        int
	Width         int
	Height        int
	Y             int
	open          bool                     // whether a menu is currently open
	dropdownMenus map[string]*DropdownMenu // dropdown menus for each menu item
}

// NewMenuWindow creates a new MenuWindow
func NewMenuWindow(x, y, w, h int) *MenuWindow {
	mw := new(MenuWindow)
	mw.MenuItems = []MenuItem{
		{Name: "File", Action: "file", Hotkey: 'i', Enabled: true},      // Alt+i (was F)
		{Name: "Edit", Action: "edit", Hotkey: 'd', Enabled: true},      // Alt+d (was E) 
		{Name: "View", Action: "view", Hotkey: 'w', Enabled: true},      // Alt+w (was V)
		{Name: "Search", Action: "search", Hotkey: 's', Enabled: true},  // Alt+s (was S)
		{Name: "Tools", Action: "tools", Hotkey: 't', Enabled: true},    // Alt+t (was T)
		{Name: "Help", Action: "help", Hotkey: 'h', Enabled: true},      // Alt+h (was H)
	}
	mw.Active = -1 // No active menu by default
	mw.Width = w
	mw.Height = h
	mw.Y = y
	mw.open = false // Menu is closed by default
	mw.dropdownMenus = make(map[string]*DropdownMenu)

	// Initialize dropdown menus
	mw.initializeDropdownMenus()

	return mw
}

// initializeDropdownMenus sets up the dropdown menus for each main menu item
func (w *MenuWindow) initializeDropdownMenus() {
	// File menu
	fileMenu := NewDropdownMenu()
	fileMenu.SetItems([]DropdownItem{
		{Text: "New", Action: "NewTab", Hotkey: 'N', Enabled: true},
		{Text: "Open", Action: "Open", Hotkey: 'O', Enabled: true},
		{Separator: true},
		{Text: "Save", Action: "Save", Hotkey: 'S', Enabled: true},
		{Text: "Save As", Action: "SaveAs", Hotkey: 'A', Enabled: true},
		{Separator: true},
		{Text: "Quit", Action: "Quit", Hotkey: 'Q', Enabled: true},
	})
	w.dropdownMenus["file"] = fileMenu

	// Edit menu
	editMenu := NewDropdownMenu()
	editMenu.SetItems([]DropdownItem{
		{Text: "Undo", Action: "Undo", Hotkey: 'U', Enabled: true},
		{Text: "Redo", Action: "Redo", Hotkey: 'R', Enabled: true},
		{Separator: true},
		{Text: "Cut", Action: "Cut", Hotkey: 'X', Enabled: true},
		{Text: "Copy", Action: "Copy", Hotkey: 'C', Enabled: true},
		{Text: "Paste", Action: "Paste", Hotkey: 'V', Enabled: true},
	})
	w.dropdownMenus["edit"] = editMenu

	// View menu
	viewMenu := NewDropdownMenu()
	viewMenu.SetItems([]DropdownItem{
		{Text: "Split Horizontal", Action: "HSplit", Hotkey: 'H', Enabled: true},
		{Text: "Split Vertical", Action: "VSplit", Hotkey: 'V', Enabled: true},
		{Separator: true},
		{Text: "Toggle Line Numbers", Action: "ToggleRuler", Hotkey: 'L', Enabled: true},
	})
	w.dropdownMenus["view"] = viewMenu

	// Search menu
	searchMenu := NewDropdownMenu()
	searchMenu.SetItems([]DropdownItem{
		{Text: "Find", Action: "Find", Hotkey: 'F', Enabled: true},
		{Text: "Find Next", Action: "FindNext", Hotkey: 'N', Enabled: true},
		{Text: "Find Previous", Action: "FindPrevious", Hotkey: 'P', Enabled: true},
		{Separator: true},
		{Text: "Replace", Action: "Replace", Hotkey: 'R', Enabled: true},
	})
	w.dropdownMenus["search"] = searchMenu

	// Tools menu
	toolsMenu := NewDropdownMenu()
	toolsMenu.SetItems([]DropdownItem{
		{Text: "Command Palette", Action: "CommandMode", Hotkey: 'C', Enabled: true},
		{Text: "Plugin Manager", Action: "PluginInstall", Hotkey: 'P', Enabled: true},
	})
	w.dropdownMenus["tools"] = toolsMenu

	// Help menu
	helpMenu := NewDropdownMenu()
	helpMenu.SetItems([]DropdownItem{
		{Text: "Help", Action: "ToggleHelp", Hotkey: 'H', Enabled: true},
		{Text: "Key Bindings", Action: "ShowKey", Hotkey: 'K', Enabled: true},
		{Separator: true},
		{Text: "About", Action: "ShowAbout", Hotkey: 'A', Enabled: true},
	})
	w.dropdownMenus["help"] = helpMenu
}

// Resize adjusts the menu window size
func (w *MenuWindow) Resize(width, height int) {
	w.Width = width
}

// SetActive sets the active menu item
func (w *MenuWindow) SetActive(index int) {
	if index >= 0 && index < len(w.MenuItems) {
		w.Active = index
	} else {
		w.Active = -1
	}
}

// GetActive returns the currently active menu item
func (w *MenuWindow) GetActive() int {
	return w.Active
}

// IsOpen returns whether a menu is currently open
func (w *MenuWindow) IsOpen() bool {
	return w.open
}

// SetOpen sets the menu open state
func (w *MenuWindow) SetOpen(open bool) {
	w.open = open

	// Show/hide the appropriate dropdown menu
	if open && w.Active >= 0 && w.Active < len(w.MenuItems) {
		activeItem := w.MenuItems[w.Active]
		if dropdown, exists := w.dropdownMenus[activeItem.Action]; exists {
			// Calculate dropdown position
			dropdownX := w.getMenuItemX(w.Active)
			dropdownY := w.Y + 1 // Below the menu bar
			dropdown.Show(dropdownX, dropdownY)
		}
	} else {
		// Hide all dropdown menus
		for _, dropdown := range w.dropdownMenus {
			dropdown.Hide()
		}
	}
}

// getMenuItemX calculates the X position of a menu item
func (w *MenuWindow) getMenuItemX(index int) int {
	x := 0
	for i := 0; i < index && i < len(w.MenuItems); i++ {
		item := w.MenuItems[i]
		if !item.Enabled {
			continue
		}
		itemWidth := util.StringWidth([]byte(item.Name), util.CharacterCountInString(item.Name), 1)
		x += itemWidth + 2 // +2 for padding
	}
	return x
}

// Display renders the menu bar
func (w *MenuWindow) Display() {
	if w.Height <= 0 {
		return
	}

	// Clear the menu bar area
	for x := 0; x < w.Width; x++ {
		screen.SetContent(x, w.Y, ' ', nil, config.DefStyle)
	}

	x := 0
	for i, item := range w.MenuItems {
		if !item.Enabled {
			continue
		}

		// Calculate item display text
		displayText := item.Name
		itemWidth := util.StringWidth([]byte(displayText), util.CharacterCountInString(displayText), 1)

		// Add padding
		padding := 2
		totalWidth := itemWidth + padding

		// Check if we have space for this item
		if x+totalWidth > w.Width {
			break
		}

		// Determine style based on active state
		style := config.DefStyle
		if i == w.Active {
			// Highlight active menu item
			style = style.Reverse(true)
		}

		// Add left padding
		screen.SetContent(x, w.Y, ' ', nil, style)
		x++

		// Render the menu item text with hotkey highlighting
		for j, r := range displayText {
			charStyle := style
			// Highlight the hotkey character
			if r == item.Hotkey || (r >= 'A' && r <= 'Z' && r-'A'+'a' == item.Hotkey) {
				charStyle = charStyle.Underline(true)
			}

			screen.SetContent(x, w.Y, r, nil, charStyle)
			x += runewidth.RuneWidth(r)

			// Handle zero-width characters
			if runewidth.RuneWidth(r) == 0 && j > 0 {
				x = x - 1
			}
		}

		// Add right padding
		screen.SetContent(x, w.Y, ' ', nil, style)
		x++
	}

	// Fill remaining space with default style
	for x < w.Width {
		screen.SetContent(x, w.Y, ' ', nil, config.DefStyle)
		x++
	}

	// Note: Dropdown menus are now displayed separately in the main event loop
	// to ensure they appear on top of all other content
}

// HandleClick handles mouse clicks on the menu bar and dropdowns
func (w *MenuWindow) HandleClick(x, y int) *DropdownItem {
	// First check if click is on an open dropdown
	if w.open && w.Active >= 0 && w.Active < len(w.MenuItems) {
		activeItem := w.MenuItems[w.Active]
		if dropdown, exists := w.dropdownMenus[activeItem.Action]; exists && dropdown.IsVisible() {
			if clickedItem := dropdown.HandleClick(x, y); clickedItem != nil {
				// A dropdown item was clicked - return it for execution
				w.SetActive(-1)
				w.SetOpen(false)
				return clickedItem
			}
			// Click might have closed the dropdown, check if we should handle menu bar click
			if !dropdown.IsVisible() {
				w.SetActive(-1)
				w.SetOpen(false)
			}
		}
	}

	// Check if click is on menu bar
	if y != w.Y {
		// Click outside menu bar and dropdown - close any open menu
		if w.open {
			w.SetActive(-1)
			w.SetOpen(false)
		}
		return nil
	}

	// Calculate which menu item was clicked
	currentX := 0
	for i, item := range w.MenuItems {
		if !item.Enabled {
			continue
		}

		itemWidth := util.StringWidth([]byte(item.Name), util.CharacterCountInString(item.Name), 1) + 2 // +2 for padding

		if x >= currentX && x < currentX+itemWidth {
			if w.Active == i && w.open {
				// Close if clicking on already open menu
				w.SetActive(-1)
				w.SetOpen(false)
			} else {
				// Activate and open menu
				w.SetActive(i)
				w.SetOpen(true)
			}
			return nil
		}

		currentX += itemWidth
	}

	// Click outside menu items - close any open menu
	w.SetActive(-1)
	w.SetOpen(false)
	return nil
}

// HandleKey handles keyboard input for menu navigation
func (w *MenuWindow) HandleKey(key rune) bool {
	// Check for hotkey matches
	for i, item := range w.MenuItems {
		if !item.Enabled {
			continue
		}

		if key == item.Hotkey || (key >= 'A' && key <= 'Z' && key-'A'+'a' == item.Hotkey) {
			w.SetActive(i)
			w.SetOpen(true)
			return true
		}
	}

	return false
}

// HandleKeyNavigation handles keyboard navigation for menu and dropdown
func (w *MenuWindow) HandleKeyNavigation(key rune, keyCode int) *DropdownItem {
	// If no menu is active, check for Alt+hotkey combinations
	if !w.open || w.Active < 0 {
		// Check for hotkey matches to open menus
		for i, item := range w.MenuItems {
			if !item.Enabled {
				continue
			}

			if key == item.Hotkey || (key >= 'A' && key <= 'Z' && key-'A'+'a' == item.Hotkey) {
				w.SetActive(i)
				w.SetOpen(true)
				return nil
			}
		}
		return nil
	}

	// If a menu is open, handle dropdown navigation
	if w.Active >= 0 && w.Active < len(w.MenuItems) {
		activeItem := w.MenuItems[w.Active]
		if dropdown, exists := w.dropdownMenus[activeItem.Action]; exists && dropdown.IsVisible() {
			// Use tcell key constants for proper key detection
			switch keyCode {
			case int(tcell.KeyEnter):
				selectedItem := dropdown.GetActiveItem()
				if selectedItem != nil && selectedItem.Enabled && !selectedItem.Separator {
					w.SetActive(-1)
					w.SetOpen(false)
					return selectedItem
				}
			case int(tcell.KeyEscape):
				w.SetActive(-1)
				w.SetOpen(false)
				return nil
			case int(tcell.KeyUp):
				dropdown.MoveUp()
				return nil
			case int(tcell.KeyDown):
				dropdown.MoveDown()
				return nil
			case int(tcell.KeyLeft):
				w.navigateToPreviousMenu()
				return nil
			case int(tcell.KeyRight):
				w.navigateToNextMenu()
				return nil
			default:
				// Check for dropdown item hotkeys
				for _, item := range dropdown.Items {
					if !item.Separator && item.Enabled {
						if key == item.Hotkey || (key >= 'A' && key <= 'Z' && key-'A'+'a' == item.Hotkey) {
							w.SetActive(-1)
							w.SetOpen(false)
							return &item
						}
					}
				}
			}
		}
	}

	return nil
}

// navigateToPreviousMenu moves to the previous menu item
func (w *MenuWindow) navigateToPreviousMenu() {
	if w.Active <= 0 {
		// Wrap to last menu
		for i := len(w.MenuItems) - 1; i >= 0; i-- {
			if w.MenuItems[i].Enabled {
				w.SetActive(i)
				w.SetOpen(true)
				return
			}
		}
	} else {
		// Move to previous enabled menu
		for i := w.Active - 1; i >= 0; i-- {
			if w.MenuItems[i].Enabled {
				w.SetActive(i)
				w.SetOpen(true)
				return
			}
		}
	}
}

// navigateToNextMenu moves to the next menu item
func (w *MenuWindow) navigateToNextMenu() {
	if w.Active >= len(w.MenuItems)-1 {
		// Wrap to first menu
		for i := 0; i < len(w.MenuItems); i++ {
			if w.MenuItems[i].Enabled {
				w.SetActive(i)
				w.SetOpen(true)
				return
			}
		}
	} else {
		// Move to next enabled menu
		for i := w.Active + 1; i < len(w.MenuItems); i++ {
			if w.MenuItems[i].Enabled {
				w.SetActive(i)
				w.SetOpen(true)
				return
			}
		}
	}
}

// GetMenuAction returns the action for the currently active menu
func (w *MenuWindow) GetMenuAction() string {
	if w.Active >= 0 && w.Active < len(w.MenuItems) {
		return w.MenuItems[w.Active].Action
	}
	return ""
}

// GetActiveDropdown returns the currently active dropdown menu
func (w *MenuWindow) GetActiveDropdown() *DropdownMenu {
	if w.open && w.Active >= 0 && w.Active < len(w.MenuItems) {
		activeItem := w.MenuItems[w.Active]
		if dropdown, exists := w.dropdownMenus[activeItem.Action]; exists {
			return dropdown
		}
	}
	return nil
}
