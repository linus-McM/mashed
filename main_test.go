package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
)

// menuItem describes an expected menu item for table-driven tests.
type menuItem struct {
	label       string
	isSeparator bool
	accelerator *keys.Accelerator
}

// assertMenuItems validates a slice of menu items against expected definitions.
func assertMenuItems(t *testing.T, items []*menu.MenuItem, expected []menuItem) {
	t.Helper()
	require.Len(t, items, len(expected), "submenu item count mismatch")

	for i, exp := range expected {
		t.Run(exp.label, func(t *testing.T) {
			item := items[i]
			if exp.isSeparator {
				assert.Equal(t, menu.SeparatorType, item.Type, "item %d should be a separator", i)
				return
			}
			assert.Equal(t, exp.label, item.Label, "item %d label mismatch", i)
			if exp.accelerator == nil {
				assert.Nil(t, item.Accelerator, "item %d should have no accelerator", i)
			} else {
				require.NotNil(t, item.Accelerator, "item %d must have an accelerator", i)
				assert.Equal(t, exp.accelerator.Key, item.Accelerator.Key, "item %d accelerator key mismatch", i)
				assert.ElementsMatch(t, exp.accelerator.Modifiers, item.Accelerator.Modifiers, "item %d modifiers mismatch", i)
			}
		})
	}
}

func TestBuildMenu_AC1_FiveMenus(t *testing.T) {
	app := &App{}
	m := buildMenu(app)
	require.NotNil(t, m, "buildMenu must return a non-nil menu")
	require.Len(t, m.Items, 5, "menu bar must have exactly 5 top-level items")

	tests := []struct {
		index int
		label string
		role  menu.Role
	}{
		{0, "mashed", 0},
		{1, "File", 0},
		{2, "", menu.EditMenuRole}, // Edit is a role menu; label may vary
		{3, "View", 0},
		{4, "Help", 0},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			item := m.Items[tt.index]
			if tt.role != 0 {
				assert.Equal(t, tt.role, item.Role, "item %d should have EditMenuRole", tt.index)
			} else {
				assert.Equal(t, tt.label, item.Label, "item %d label mismatch", tt.index)
			}
		})
	}
}

func TestBuildMenu_AC1_MashedSubmenu(t *testing.T) {
	app := &App{}
	m := buildMenu(app)
	require.NotNil(t, m)
	require.GreaterOrEqual(t, len(m.Items), 1)

	mashedItem := m.Items[0]
	require.Equal(t, "mashed", mashedItem.Label)
	require.NotNil(t, mashedItem.SubMenu, "mashed must have a submenu")

	expected := []menuItem{
		{label: "About mashed", accelerator: nil},
		{isSeparator: true},
		{label: "Settings...", accelerator: keys.CmdOrCtrl(",")},
		{isSeparator: true},
		{label: "Quit mashed", accelerator: keys.CmdOrCtrl("q")},
	}

	assertMenuItems(t, mashedItem.SubMenu.Items, expected)
}

func TestBuildMenu_AC2_FileSubmenu(t *testing.T) {
	app := &App{}
	m := buildMenu(app)
	require.NotNil(t, m)
	require.GreaterOrEqual(t, len(m.Items), 2)

	fileItem := m.Items[1]
	require.Equal(t, "File", fileItem.Label)
	require.NotNil(t, fileItem.SubMenu, "File must have a submenu")

	expected := []menuItem{
		{label: "New Agent", accelerator: keys.CmdOrCtrl("n")},
		{label: "New Repository...", accelerator: keys.Combo("n", keys.CmdOrCtrlKey, keys.ShiftKey)},
		{isSeparator: true},
		{label: "Open Workspace...", accelerator: keys.CmdOrCtrl("o")},
		{isSeparator: true},
		{label: "Take Screenshot", accelerator: keys.Combo("s", keys.CmdOrCtrlKey, keys.ShiftKey)},
	}

	assertMenuItems(t, fileItem.SubMenu.Items, expected)
}

func TestBuildMenu_AC4_ViewSubmenu(t *testing.T) {
	app := &App{}
	m := buildMenu(app)
	require.NotNil(t, m)
	require.GreaterOrEqual(t, len(m.Items), 4)

	viewItem := m.Items[3]
	require.Equal(t, "View", viewItem.Label)
	require.NotNil(t, viewItem.SubMenu, "View must have a submenu")

	expected := []menuItem{
		{label: "Feed", accelerator: keys.CmdOrCtrl("1")},
		{label: "Workflows", accelerator: keys.CmdOrCtrl("2")},
		{isSeparator: true},
		{label: "Minimize", accelerator: keys.CmdOrCtrl("m")},
		{label: "Toggle Fullscreen", accelerator: keys.Combo("f", keys.ControlKey, keys.CmdOrCtrlKey)},
	}

	assertMenuItems(t, viewItem.SubMenu.Items, expected)
}

func TestBuildMenu_AC5_HelpSubmenu(t *testing.T) {
	app := &App{}
	m := buildMenu(app)
	require.NotNil(t, m)
	require.GreaterOrEqual(t, len(m.Items), 5)

	helpItem := m.Items[4]
	require.Equal(t, "Help", helpItem.Label)
	require.NotNil(t, helpItem.SubMenu, "Help must have a submenu")

	items := helpItem.SubMenu.Items
	require.Len(t, items, 1, "Help submenu must have 1 item")

	assert.Equal(t, "mashed on GitHub", items[0].Label)
	assert.Nil(t, items[0].Accelerator, "mashed on GitHub should have no accelerator")
}
