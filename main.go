package main

/*
#cgo CXXFLAGS: -std=c++11
#cgo LDFLAGS: -L. wrapper.o -lstdc++ -ltag
#include "wrapper.h"
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"os"
	"unsafe"
	"path/filepath"
	"strings"
	
	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var log = ""

var supportedExtensions = map[string]bool{
	".mp3":  true,
	".m4a":  true,
	".flac": true,
	".ogg":  true,
	".wav":  true,
	".aiff": true,
	".wma":  true,
	".aac":  true,
	".opus": true,
	".ape":  true,
	".wv":   true,
	".mpc":  true,
}

type state int

type Metadata struct {
	Title						string
	Artist					string
	AlbumArtist			string
	Album						string
	Genre						string
	Comment					string
	Codec						string
	TagType					string
	ReleaseDate			string
	Year						int
	DiscNumber			int
	DiscTotal				int
	TrackNumber			int
	TrackTotal			int
	Bitrate					int // in kb/s
	Frequency				int // in Hz
	Duration				int // in seconds
	Channels				int
	HasImage				bool
}

const (
	filePickerState state = iota
	metadataTableState
	editingState
)

type model struct {
	state          state
	filepicker     filepicker.Model
	table          table.Model
	textInput      textinput.Model
	selectedFile   string
	editingRow     int
	editingProperty string
	metadata       map[string]string
	width          int
	height         int
}

type fileSelectedMsg struct {
	path string
}

func initialModel() model {
	fp := filepicker.New()
	fp.AllowedTypes = []string{".mp3", ".m4a", ".flac", ".ogg", ".wav", ".aiff", ".wma", ".aac", ".opus", ".ape", ".wv", ".mpc"}
	fp.CurrentDirectory, _ = os.Getwd()

	ti := textinput.New()
	ti.Placeholder = "Enter new value..."
	ti.CharLimit = 256
	ti.Width = 50

	return model{
		state:      filePickerState,
		filepicker: fp,
		textInput:  ti,
		metadata:   make(map[string]string),
		width:      80,
		height:     24,
	}
}

func EditMetadata(filename, field, value string) error {
	cFname := C.CString(filename)
	cField := C.CString(field)
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cFname))
	defer C.free(unsafe.Pointer(cField))
	defer C.free(unsafe.Pointer(cValue))

	success := C.edit_metadata(cFname, cField, cValue)

	if int(success) == -1 {
		return fmt.Errorf("Error writing metadata")
	}
	return nil
}

func (m model) Init() tea.Cmd {
	return m.filepicker.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	/*
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.state == metadataTableState {
			m.table.SetWidth(msg.Width - 4)
			m.table.SetHeight(msg.Height - 8)
		}
		return m, nil
	*/
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			if m.state == metadataTableState {
				m.state = filePickerState
				return m, nil
			} else if m.state == editingState {
				m.state = metadataTableState
				m.textInput.Blur()
				return m, nil
			}
		case "enter":
			if m.state == metadataTableState {
				selectedRow := m.table.Cursor()
				if selectedRow < len(m.table.Rows()) {
					row := m.table.Rows()[selectedRow]
					property := row[0]

					// Only allow editing of editable properties
					if m.isEditableProperty(property) {
						m.editingRow = selectedRow
						m.editingProperty = property
						m.textInput.SetValue(row[1])
						m.textInput.Focus()
						m.state = editingState
						return m, textinput.Blink
					}
				}
			} else if m.state == editingState {
				// Save the edited value
				newValue := m.textInput.Value()
				if err := EditMetadata(m.selectedFile, strings.ToLower(m.editingProperty), newValue); err != nil {
					return m, nil
				}
				m.metadata[m.editingProperty] = newValue
				m.updateTableRow(m.editingRow, m.editingProperty, newValue)
				m.extractMetadata(m.selectedFile)
				m.textInput.Blur()
				m.state = metadataTableState
				return m, nil
			}
		}

	case fileSelectedMsg:
		m.selectedFile = msg.path
		m.state = metadataTableState
		m.extractMetadata(msg.path)
		m.table = m.createMetadataTable()
		return m, nil
	}

	var cmd tea.Cmd
	switch m.state {
	case filePickerState:
		m.filepicker, cmd = m.filepicker.Update(msg)
		if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
			if isAudioFile(path) {
				return m, func() tea.Msg {
					return fileSelectedMsg{path: path}
				}
			}
		}
	case metadataTableState:
		m.table, cmd = m.table.Update(msg)
	case editingState:
		m.textInput, cmd = m.textInput.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	var s strings.Builder

	switch m.state {
	case filePickerState:
		s.WriteString("Select an audio file to edit metadata:\n\n")
		s.WriteString(m.filepicker.View())
		s.WriteString("\n\nPress q to quit")
	case metadataTableState:
		s.WriteString(fmt.Sprintf("Metadata for: %s\n\n", filepath.Base(m.selectedFile)))
		s.WriteString(m.table.View())
		s.WriteString("\n\nPress ENTER to edit selected property, ESC to go back, q to quit")
	case editingState:
		s.WriteString(fmt.Sprintf("Editing: %s\n\n", m.editingProperty))
		s.WriteString(m.textInput.View())
		s.WriteString("\n\nPress ENTER to save, ESC to cancel")
	}

	return s.String()
}

func (m model) createMetadataTable() table.Model {
	columns := []table.Column{
		{Title: "Property", Width: 20},
		{Title: "Value", Width: 50},
	}

	rows := m.buildTableRows()

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(m.height-8),
		table.WithWidth(m.width-4),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return t
}

func (m *model) extractMetadata(filePath string) {
	// Initialize metadata map
	m.metadata = make(map[string]string)

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		m.metadata["Error"] = err.Error()
		return
	}

	ext := strings.ToLower(filepath.Ext(filePath))

	// Set basic file info (non-editable)
	m.metadata["File Name"] = filepath.Base(filePath)
	m.metadata["File Path"] = filePath
	m.metadata["File Size"] = formatFileSize(fileInfo.Size())
	m.metadata["Modified"] = fileInfo.ModTime().Format("2006-01-02 15:04:05")
	m.metadata["Format"] = strings.TrimPrefix(ext, ".")

	// Add editable metadata based on file type
	meta, err := ReadMetadata(filePath)
	if err != nil {
		m.metadata["Error"] = fmt.Sprintf("Failed to read metadata: %v", err)
		return
	}

	// Set the actual metadata values
	if meta.Title != "" {
		m.metadata["Title"] = meta.Title
	} else {
		m.metadata["Title"] = "Unknown"
	}

	if meta.Artist != "" {
		m.metadata["Artist"] = meta.Artist
	} else {
		m.metadata["Artist"] = "Unknown"
	}

	if meta.Album != "" {
		m.metadata["Album"] = meta.Album
	} else {
		m.metadata["Album"] = "Unknown"
	}

	// Add additional metadata if available
	if meta.Year > 0 {
		m.metadata["Year"] = fmt.Sprintf("%d", meta.Year)
	}
	if meta.Genre != "" {
		m.metadata["Genre"] = meta.Genre
	}
	if meta.TrackNumber > 0 {
		if meta.TrackTotal > 0 {
			m.metadata["Track"] = fmt.Sprintf("%d/%d", meta.TrackNumber, meta.TrackTotal)
		} else {
			m.metadata["Track"] = fmt.Sprintf("%d", meta.TrackNumber)
		}
	}
	if meta.Bitrate > 0 {
		m.metadata["Bitrate"] = fmt.Sprintf("%d kbps", meta.Bitrate)
	}
	if meta.Duration > 0 {
		minutes := meta.Duration / 60
		seconds := meta.Duration % 60
		m.metadata["Duration"] = fmt.Sprintf("%d:%02d", minutes, seconds)
	}
	if meta.Frequency > 0 {
		m.metadata["Sample Rate"] = fmt.Sprintf("%d Hz", meta.Frequency)
	}
	if meta.Codec != "" {
		m.metadata["Codec"] = meta.Codec
	}

}

func (m model) buildTableRows() []table.Row {
	// Define the order of properties to display
	propertyOrder := []string{
		"File Name", "File Path", "File Size", "Modified", "Format",
		"Title", "Artist", "Album", "Year", "Date", "Genre", "Track", "Track Number",
		"Bitrate", "Duration", "Sample Rate", "Bit Depth", "ID3 Version", "Codec",
	}

	var rows []table.Row

	// Add properties in order if they exist
	for _, prop := range propertyOrder {
		if value, exists := m.metadata[prop]; exists {
			rows = append(rows, table.Row{prop, value})
		}
	}

	// Add any remaining properties not in the order
	for prop, value := range m.metadata {
		found := false
		for _, orderedProp := range propertyOrder {
			if prop == orderedProp {
				found = true
				break
			}
		}
		if !found {
			rows = append(rows, table.Row{prop, value})
		}
	}

	return rows
}

func (m model) isEditableProperty(property string) bool {
	editableProperties := map[string]bool{
		"Title":        true,
		"Artist":       true,
		"Album":        true,
		"Year":         true,
		"Date":         true,
		"Genre":        true,
		"Track":        true,
		"Track Number": true,
	}
	return editableProperties[property]
}

func (m *model) updateTableRow(rowIndex int, property string, newValue string) {
	m.metadata[property] = newValue

	// Rebuild the table with updated data
	rows := m.buildTableRows()
	m.table.SetRows(rows)
}

func isAudioFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return supportedExtensions[ext]
}

func formatFileSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d bytes", size)
	}
}

func ReadMetadata(filename string) (*Metadata, error) {
	cPath := C.CString(filename)
	defer C.free(unsafe.Pointer(cPath))

	cMeta := C.read_metadata(cPath)
	if cMeta == nil {
		return nil, fmt.Errorf("failed to read metadata")
	}
	defer C.free_metadata(cMeta)

	meta := &Metadata{
		Title:				C.GoString(cMeta.title),
		Artist:				C.GoString(cMeta.artist),
		AlbumArtist:	C.GoString(cMeta.album_artist),
		Album:				C.GoString(cMeta.album),
		Genre:				C.GoString(cMeta.genre),
		Comment:			C.GoString(cMeta.comment),
		Codec:				C.GoString(cMeta.codec),
		TagType:			C.GoString(cMeta.tag_type),
		ReleaseDate:	C.GoString(cMeta.date),
		Year:					int(cMeta.year),
		DiscNumber:		int(cMeta.disc),
		DiscTotal:		int(cMeta.disc_total),
		TrackNumber:	int(cMeta.track),
		TrackTotal:		int(cMeta.track_total),
		Bitrate:			int(cMeta.bitrate),
		Frequency:		int(cMeta.frequency),
		Duration:			int(cMeta.duration),
		Channels:			int(cMeta.channels),
		HasImage:			bool(cMeta.has_image>0),
	}
	return meta, nil
}

func main() {
	if len(os.Args) < 2 {
		p := tea.NewProgram(initialModel(), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	meta, err := ReadMetadata(os.Args[1])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("%+v\n", meta)
	fmt.Printf("%v\n", log)

	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
