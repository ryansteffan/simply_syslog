package drivers

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/ryansteffan/simply_syslog/internal/buffer"
	"github.com/ryansteffan/simply_syslog/internal/config"
)

type FileWriter struct {
	Name    string
	Enabled bool
	Options struct {
		Path string
	}
}

func NewFileWriter() Driver {
	return &FileWriter{
		Enabled: false,
	}
}

// GetName implements [Driver].
func (f *FileWriter) GetName() string {
	return f.Name
}

// IsEnabled implements [Driver].
func (f *FileWriter) IsEnabled() bool {
	return f.Enabled
}

// Setup implements [Driver].
func (f *FileWriter) Setup(config config.Writer) error {
	f.Name = config.Name
	f.Enabled = config.Enabled

	if path, ok := config.Options["path"]; ok {
		f.Options.Path = path
	} else {
		return errors.New("path directive not present in writer config")
	}

	// Create the log file if it does not exist.
	if _, err := os.Stat(f.Options.Path); os.IsNotExist(err) {
		err := os.MkdirAll(filepath.Dir(f.Options.Path), 0755)
		if err != nil {
			return err
		}
		file, err := os.Create(f.Options.Path)
		if err != nil {
			return err
		}
		file.Close()
	}

	return nil
}

// Write implements [Driver].
func (f *FileWriter) Write(data buffer.BufferTransferData) error {
	// Open the file in append mode.
	file, err := os.OpenFile(f.Options.Path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(data.RawMessage + "\n")
	if err != nil {
		return err
	}

	return nil
}

var _ Driver = (*FileWriter)(nil)
