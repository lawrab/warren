package ui

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/lawrab/warren/internal/fileops"
	"github.com/lawrab/warren/pkg/models"
)

// FileView represents the main file listing widget.
type FileView struct {
	widget        *gtk.ScrolledWindow
	listView      *gtk.ColumnView
	store         *gio.ListStore
	currentPath   string
	selectedIndex int
	files         []models.FileInfo
	showHidden    bool
	sortMode      models.SortBy
	sortOrder     models.SortOrder
}

// NewFileView creates a new file listing widget.
func NewFileView() *FileView {
	fv := &FileView{
		selectedIndex: -1,
		showHidden:    false,
		files:         make([]models.FileInfo, 0),
		sortMode:      models.SortByName,
		sortOrder:     models.SortAscending,
	}

	// Create list store to hold file data
	fv.store = gio.NewListStore(glib.TypeObject)

	// Create selection model
	selection := gtk.NewSingleSelection(fv.store)
	selection.SetAutoselect(false)
	selection.SetCanUnselect(true)

	// Create column view
	fv.listView = gtk.NewColumnView(selection)

	// Add columns
	fv.addColumns()

	// Create scrolled window
	fv.widget = gtk.NewScrolledWindow()
	fv.widget.SetChild(fv.listView)
	fv.widget.SetVExpand(true)
	fv.widget.SetHExpand(true)

	return fv
}

// addColumns adds the columns to the column view (name, size, modified).
func (fv *FileView) addColumns() {
	// Name column
	nameFactory := gtk.NewSignalListItemFactory()
	nameFactory.ConnectSetup(func(obj *glib.Object) {
		cell := obj.Cast().(*gtk.ColumnViewCell)
		label := gtk.NewLabel("")
		label.SetXAlign(0) // Left align
		cell.SetChild(label)
	})
	nameFactory.ConnectBind(func(obj *glib.Object) {
		cell := obj.Cast().(*gtk.ColumnViewCell)
		label := cell.Child().(*gtk.Label)

		// Get the file info from the position
		pos := cell.Position()
		if pos < uint(len(fv.files)) {
			file := fv.files[pos]
			icon := "📄"
			if file.IsDir {
				icon = "📁"
			} else if file.IsSymlink {
				icon = "🔗"
			}
			label.SetText(fmt.Sprintf("%s %s", icon, file.Name))
		}
	})

	nameColumn := gtk.NewColumnViewColumn("Name", &nameFactory.ListItemFactory)
	nameColumn.SetExpand(true)
	fv.listView.AppendColumn(nameColumn)

	// Size column
	sizeFactory := gtk.NewSignalListItemFactory()
	sizeFactory.ConnectSetup(func(obj *glib.Object) {
		cell := obj.Cast().(*gtk.ColumnViewCell)
		label := gtk.NewLabel("")
		label.SetXAlign(1) // Right align
		cell.SetChild(label)
	})
	sizeFactory.ConnectBind(func(obj *glib.Object) {
		cell := obj.Cast().(*gtk.ColumnViewCell)
		label := cell.Child().(*gtk.Label)

		pos := cell.Position()
		if pos < uint(len(fv.files)) {
			file := fv.files[pos]
			if file.IsDir {
				label.SetText("-")
			} else {
				label.SetText(fileops.FormatSize(file.Size))
			}
		}
	})

	sizeColumn := gtk.NewColumnViewColumn("Size", &sizeFactory.ListItemFactory)
	sizeColumn.SetFixedWidth(100)
	fv.listView.AppendColumn(sizeColumn)

	// Modified column
	modFactory := gtk.NewSignalListItemFactory()
	modFactory.ConnectSetup(func(obj *glib.Object) {
		cell := obj.Cast().(*gtk.ColumnViewCell)
		label := gtk.NewLabel("")
		label.SetXAlign(0)
		cell.SetChild(label)
	})
	modFactory.ConnectBind(func(obj *glib.Object) {
		cell := obj.Cast().(*gtk.ColumnViewCell)
		label := cell.Child().(*gtk.Label)

		pos := cell.Position()
		if pos < uint(len(fv.files)) {
			file := fv.files[pos]
			label.SetText(formatModTime(file.ModTime))
		}
	})

	modColumn := gtk.NewColumnViewColumn("Modified", &modFactory.ListItemFactory)
	modColumn.SetFixedWidth(150)
	fv.listView.AppendColumn(modColumn)
}

// Widget returns the GTK widget.
func (fv *FileView) Widget() gtk.Widgetter {
	return fv.widget
}

// LoadDirectory loads and displays the contents of a directory.
func (fv *FileView) LoadDirectory(path string) error {
	files, err := fileops.ListDirectory(path, fv.showHidden)
	if err != nil {
		return fmt.Errorf("failed to load directory: %w", err)
	}

	// Sort files using current sort mode and order
	fileops.SortFiles(files, fv.sortMode, fv.sortOrder)

	fv.files = files
	fv.currentPath = path

	// Clear the store
	fv.store.RemoveAll()

	// Add files to store (we use StringObject as placeholders)
	for i := range files {
		obj := gtk.NewStringObject(fmt.Sprintf("%d", i))
		fv.store.Append(obj.Object)
	}

	// Reset selection
	fv.selectedIndex = -1
	if len(files) > 0 {
		fv.SelectIndex(0)
	}

	return nil
}

// SelectIndex selects the file at the given index.
func (fv *FileView) SelectIndex(index int) {
	if index < 0 || index >= len(fv.files) {
		return
	}

	fv.selectedIndex = index
	model := fv.listView.Model()
	selection := model.Cast().(*gtk.SingleSelection)
	selection.SetSelected(uint(index))
}

// SelectNext moves selection down one item.
func (fv *FileView) SelectNext() {
	if fv.selectedIndex < len(fv.files)-1 {
		fv.SelectIndex(fv.selectedIndex + 1)
	}
}

// SelectPrevious moves selection up one item.
func (fv *FileView) SelectPrevious() {
	if fv.selectedIndex > 0 {
		fv.SelectIndex(fv.selectedIndex - 1)
	}
}

// GetSelected returns the currently selected file, or nil if none selected.
func (fv *FileView) GetSelected() *models.FileInfo {
	if fv.selectedIndex < 0 || fv.selectedIndex >= len(fv.files) {
		return nil
	}
	return &fv.files[fv.selectedIndex]
}

// GetCurrentPath returns the current directory path.
func (fv *FileView) GetCurrentPath() string {
	return fv.currentPath
}

// NavigateUp navigates to the parent directory.
func (fv *FileView) NavigateUp() error {
	parent := fileops.GetParentDir(fv.currentPath)
	if parent == fv.currentPath {
		// Already at root
		return nil
	}
	return fv.LoadDirectory(parent)
}

// NavigateInto enters the selected directory.
func (fv *FileView) NavigateInto() error {
	selected := fv.GetSelected()
	if selected == nil {
		return fmt.Errorf("no file selected")
	}

	if !selected.IsDir {
		return fmt.Errorf("not a directory")
	}

	return fv.LoadDirectory(selected.Path)
}

// ToggleHidden toggles the visibility of hidden files.
func (fv *FileView) ToggleHidden() error {
	fv.showHidden = !fv.showHidden
	return fv.LoadDirectory(fv.currentPath)
}

// formatModTime formats a time for display in the file list.
func formatModTime(t time.Time) string {
	now := time.Now()
	if t.Year() == now.Year() {
		return t.Format("Jan 02 15:04")
	}
	return t.Format("Jan 02  2006")
}

// GetFileCount returns the number of files currently displayed.
func (fv *FileView) GetFileCount() int {
	return len(fv.files)
}

// GetSelectedPath returns the path of the selected file, or empty string.
func (fv *FileView) GetSelectedPath() string {
	selected := fv.GetSelected()
	if selected == nil {
		return ""
	}
	return selected.Path
}

// ParentPath returns the parent directory of the current path.
func (fv *FileView) ParentPath() string {
	if fv.currentPath == "" {
		return ""
	}
	return filepath.Dir(fv.currentPath)
}

// SetSortMode sets the sort mode and order for the file view.
func (fv *FileView) SetSortMode(mode models.SortBy, order models.SortOrder) {
	fv.sortMode = mode
	fv.sortOrder = order
}

// CycleSortMode cycles through the available sort modes.
// Order: Name -> Size -> Modified -> Extension -> (repeat)
func (fv *FileView) CycleSortMode() error {
	switch fv.sortMode {
	case models.SortByName:
		fv.sortMode = models.SortBySize
	case models.SortBySize:
		fv.sortMode = models.SortByModTime
	case models.SortByModTime:
		fv.sortMode = models.SortByExtension
	case models.SortByExtension:
		fv.sortMode = models.SortByName
	default:
		fv.sortMode = models.SortByName
	}

	// Re-sort and refresh the display
	return fv.LoadDirectory(fv.currentPath)
}

// GetSortMode returns the current sort mode.
func (fv *FileView) GetSortMode() models.SortBy {
	return fv.sortMode
}

// GetSortOrder returns the current sort order.
func (fv *FileView) GetSortOrder() models.SortOrder {
	return fv.sortOrder
}

// ToggleSortOrder toggles between ascending and descending sort order.
func (fv *FileView) ToggleSortOrder() error {
	if fv.sortOrder == models.SortAscending {
		fv.sortOrder = models.SortDescending
	} else {
		fv.sortOrder = models.SortAscending
	}

	// Re-sort and refresh the display
	return fv.LoadDirectory(fv.currentPath)
}
