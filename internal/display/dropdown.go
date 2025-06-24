package display

import (
	runewidth "github.com/mattn/go-runewidth"
	"github.com/zyedidia/micro/v2/internal/config"
	"github.com/zyedidia/micro/v2/internal/screen"
	"github.com/zyedidia/micro/v2/internal/util"
)

// DropdownItem represents a single item in a dropdown menu
type DropdownItem struct {
	Text      string
	Action    string
	Hotkey    rune
	Enabled   bool
	Separator bool // True for separator lines
}

// DropdownMenu represents a dropdown menu that appears below menu items
type DropdownMenu struct {
	Items   []DropdownItem
	X       int
	Y       int
	Width   int
	Height  int
	Active  int  // Currently highlighted item (-1 for none)
	Visible bool // Whether the dropdown is currently shown
}

// NewDropdownMenu creates a new dropdown menu
func NewDropdownMenu() *DropdownMenu {
	return &DropdownMenu{
		Items:   []DropdownItem{},
		Active:  -1,
		Visible: false,
	}
}

// SetItems sets the items for this dropdown menu
func (d *DropdownMenu) SetItems(items []DropdownItem) {
	d.Items = items
	d.calculateSize()
}

// calculateSize determines the width and height needed for the dropdown
func (d *DropdownMenu) calculateSize() {
	d.Width = 0
	d.Height = len(d.Items) + 2 // +2 for top and bottom borders

	// Find the widest item
	for _, item := range d.Items {
		if item.Separator {
			continue
		}
		itemWidth := util.StringWidth([]byte(item.Text), util.CharacterCountInString(item.Text), 1)
		if item.Hotkey != 0 {
			itemWidth += 4 // Space for " (X)" hotkey display
		}
		if itemWidth > d.Width {
			d.Width = itemWidth
		}
	}

	// Add padding and border
	d.Width += 4 // 2 for borders + 2 for padding
	if d.Width < 8 {
		d.Width = 8 // Minimum width
	}
}

// Show displays the dropdown at the specified position
func (d *DropdownMenu) Show(x, y int) {
	d.X = x
	d.Y = y
	d.Visible = true

	// Set the first selectable item as active
	d.Active = -1
	for i := 0; i < len(d.Items); i++ {
		if d.Items[i].Enabled && !d.Items[i].Separator {
			d.Active = i
			break
		}
	}
}

// Hide hides the dropdown
func (d *DropdownMenu) Hide() {
	d.Visible = false
	d.Active = -1
}

// IsVisible returns whether the dropdown is currently visible
func (d *DropdownMenu) IsVisible() bool {
	return d.Visible
}

// Display renders the dropdown menu
func (d *DropdownMenu) Display() {
	if !d.Visible || d.Height == 0 {
		return
	}

	// Get terminal size to ensure we don't draw outside bounds
	termWidth, termHeight := screen.Screen.Size()

	// Adjust position if dropdown would go off screen
	adjustedX := d.X
	adjustedY := d.Y

	if adjustedX+d.Width > termWidth {
		adjustedX = termWidth - d.Width
		if adjustedX < 0 {
			adjustedX = 0
		}
	}

	if adjustedY+d.Height > termHeight {
		adjustedY = termHeight - d.Height
		if adjustedY < 0 {
			adjustedY = 0
		}
	}

	// Draw dropdown background and border with proper backdrop
	// Use normal style for dropdown, reverse for highlighting
	dropdownStyle := config.DefStyle
	borderStyle := config.DefStyle
	shadowStyle := config.DefStyle.Dim(true) // For drop shadow effect

	// Draw shadow effect first (offset by 1 pixel)
	for row := 1; row <= d.Height; row++ {
		for col := 1; col <= d.Width; col++ {
			x := adjustedX + col
			y := adjustedY + row
			if x < termWidth && y < termHeight {
				screen.SetContent(x, y, ' ', nil, shadowStyle)
			}
		}
	}

	for row := 0; row < d.Height; row++ {
		y := adjustedY + row
		if y >= termHeight {
			break
		}

		for col := 0; col < d.Width; col++ {
			x := adjustedX + col
			if x >= termWidth {
				break
			}

			// Draw border
			if col == 0 || col == d.Width-1 {
				if row == 0 {
					if col == 0 {
						screen.SetContent(x, y, '┌', nil, borderStyle)
					} else {
						screen.SetContent(x, y, '┐', nil, borderStyle)
					}
				} else if row == d.Height-1 {
					if col == 0 {
						screen.SetContent(x, y, '└', nil, borderStyle)
					} else {
						screen.SetContent(x, y, '┘', nil, borderStyle)
					}
				} else {
					screen.SetContent(x, y, '│', nil, borderStyle)
				}
			} else if row == 0 || row == d.Height-1 {
				screen.SetContent(x, y, '─', nil, borderStyle)
			} else {
				screen.SetContent(x, y, ' ', nil, dropdownStyle)
			}
		}
	}

	// Draw menu items
	itemY := 0
	for i, item := range d.Items {
		if itemY >= d.Height-2 { // Account for top and bottom borders
			break
		}

		y := adjustedY + 1 + itemY // +1 for top border
		if y >= termHeight {
			break
		}

		if item.Separator {
			// Draw separator line
			for x := adjustedX + 1; x < adjustedX+d.Width-1; x++ {
				if x < termWidth {
					screen.SetContent(x, y, '─', nil, borderStyle)
				}
			}
		} else {
			// Draw menu item
			itemStyle := dropdownStyle
			if i == d.Active {
				// Highlight active item
				itemStyle = itemStyle.Reverse(true)
			}
			if !item.Enabled {
				// Dim disabled items
				itemStyle = itemStyle.Dim(true)
			}

			// Clear the line first
			for x := adjustedX + 1; x < adjustedX+d.Width-1; x++ {
				if x < termWidth {
					screen.SetContent(x, y, ' ', nil, itemStyle)
				}
			}

			// Draw item text
			x := adjustedX + 2 // +2 for border and padding
			for _, r := range item.Text {
				if x >= adjustedX+d.Width-2 || x >= termWidth {
					break
				}
				screen.SetContent(x, y, r, nil, itemStyle)
				x += runewidth.RuneWidth(r)
			}

			// Draw hotkey if present
			if item.Hotkey != 0 && x < adjustedX+d.Width-4 {
				hotkeyText := " (" + string(item.Hotkey) + ")"
				for _, r := range hotkeyText {
					if x >= adjustedX+d.Width-2 || x >= termWidth {
						break
					}
					screen.SetContent(x, y, r, nil, itemStyle.Dim(true))
					x += runewidth.RuneWidth(r)
				}
			}
		}
		itemY++
	}
}

// HandleClick handles mouse clicks on the dropdown
func (d *DropdownMenu) HandleClick(x, y int) *DropdownItem {
	if !d.Visible {
		return nil
	}

	// Check if click is inside dropdown bounds
	if x < d.X || x >= d.X+d.Width || y < d.Y || y >= d.Y+d.Height {
		// Click outside dropdown - hide it
		d.Hide()
		return nil
	}

	// Check if click is on border
	if x == d.X || x == d.X+d.Width-1 || y == d.Y || y == d.Y+d.Height-1 {
		return nil
	}

	// Calculate which item was clicked
	itemIndex := y - d.Y - 1 // -1 for top border
	if itemIndex >= 0 && itemIndex < len(d.Items) {
		item := &d.Items[itemIndex]
		if !item.Separator && item.Enabled {
			d.Hide()
			return item
		}
	}

	return nil
}

// HandleKey handles keyboard navigation in the dropdown
func (d *DropdownMenu) HandleKey(key rune) *DropdownItem {
	if !d.Visible {
		return nil
	}

	// Check for hotkey matches
	for _, item := range d.Items {
		if !item.Separator && item.Enabled {
			if key == item.Hotkey || (key >= 'A' && key <= 'Z' && key-'A'+'a' == item.Hotkey) {
				d.Hide()
				return &item
			}
		}
	}

	return nil
}

// NavigateUp moves selection up in the dropdown
func (d *DropdownMenu) NavigateUp() {
	if !d.Visible {
		return
	}

	for i := d.Active - 1; i >= 0; i-- {
		if !d.Items[i].Separator && d.Items[i].Enabled {
			d.Active = i
			return
		}
	}

	// Wrap to bottom
	for i := len(d.Items) - 1; i > d.Active; i-- {
		if !d.Items[i].Separator && d.Items[i].Enabled {
			d.Active = i
			return
		}
	}
}

// NavigateDown moves selection down in the dropdown
func (d *DropdownMenu) NavigateDown() {
	if !d.Visible {
		return
	}

	for i := d.Active + 1; i < len(d.Items); i++ {
		if !d.Items[i].Separator && d.Items[i].Enabled {
			d.Active = i
			return
		}
	}

	// Wrap to top
	for i := 0; i < d.Active; i++ {
		if !d.Items[i].Separator && d.Items[i].Enabled {
			d.Active = i
			return
		}
	}
}

// SelectActive returns the currently active item and hides the dropdown
func (d *DropdownMenu) SelectActive() *DropdownItem {
	if !d.Visible || d.Active < 0 || d.Active >= len(d.Items) {
		return nil
	}

	item := &d.Items[d.Active]
	if !item.Separator && item.Enabled {
		d.Hide()
		return item
	}

	return nil
}

// GetActiveItem returns the currently active item, or nil if none
func (d *DropdownMenu) GetActiveItem() *DropdownItem {
	if d.Active >= 0 && d.Active < len(d.Items) {
		return &d.Items[d.Active]
	}
	return nil
}

// MoveUp moves selection up to previous selectable item
func (d *DropdownMenu) MoveUp() {
	if d.Active < 0 {
		// No item selected, select the last selectable item
		for i := len(d.Items) - 1; i >= 0; i-- {
			if d.Items[i].Enabled && !d.Items[i].Separator {
				d.Active = i
				return
			}
		}
	} else if d.Active == 0 {
		// At first item, wrap to last selectable item
		for i := len(d.Items) - 1; i >= 0; i-- {
			if d.Items[i].Enabled && !d.Items[i].Separator {
				d.Active = i
				return
			}
		}
	} else {
		// Move to previous selectable item
		for i := d.Active - 1; i >= 0; i-- {
			if d.Items[i].Enabled && !d.Items[i].Separator {
				d.Active = i
				return
			}
		}
		// If no previous selectable item found, wrap to last
		for i := len(d.Items) - 1; i >= 0; i-- {
			if d.Items[i].Enabled && !d.Items[i].Separator {
				d.Active = i
				return
			}
		}
	}
}

// MoveDown moves selection down to next selectable item
func (d *DropdownMenu) MoveDown() {
	if d.Active < 0 {
		// No item selected, select the first selectable item
		for i := 0; i < len(d.Items); i++ {
			if d.Items[i].Enabled && !d.Items[i].Separator {
				d.Active = i
				return
			}
		}
	} else if d.Active >= len(d.Items)-1 {
		// At last item, wrap to first selectable item
		for i := 0; i < len(d.Items); i++ {
			if d.Items[i].Enabled && !d.Items[i].Separator {
				d.Active = i
				return
			}
		}
	} else {
		// Move to next selectable item
		for i := d.Active + 1; i < len(d.Items); i++ {
			if d.Items[i].Enabled && !d.Items[i].Separator {
				d.Active = i
				return
			}
		}
		// If no next selectable item found, wrap to first
		for i := 0; i < len(d.Items); i++ {
			if d.Items[i].Enabled && !d.Items[i].Separator {
				d.Active = i
				return
			}
		}
	}
}
