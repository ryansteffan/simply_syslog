package drivers

import (
	"database/sql"
	"errors"

	_ "github.com/duckdb/duckdb-go/v2"
	"github.com/ryansteffan/simply_syslog/internal/buffer"
	"github.com/ryansteffan/simply_syslog/internal/config"
)

type DuckDBWriter struct {
	Name        string
	Enabled     bool
	RegexConfig *config.RegexConfig
	Options     struct {
		Path string
	}
}

func NewDuckDBWriter() Driver {
	return &DuckDBWriter{
		Enabled: false,
	}
}

// GetName implements [Driver].
func (s *DuckDBWriter) GetName() string {
	return s.Name
}

// IsEnabled implements [Driver].
func (s *DuckDBWriter) IsEnabled() bool {
	return s.Enabled
}

// Setup implements [Driver].
func (s *DuckDBWriter) Setup(conf config.Writer) error {
	s.Name = conf.Name
	s.Enabled = conf.Enabled
	if path, ok := conf.Options["path"]; ok {
		s.Options.Path = path
	} else {
		return errors.New("path directive not present in writer config")
	}

	// Pull the table map for the database
	regexConf, err := config.GetRegexConfig()
	if err != nil {
		return err
	}
	s.RegexConfig = regexConf

	// Create the database and required tables
	sql, err := sql.Open("duckdb", s.Options.Path)
	if err != nil {
		return err
	}
	defer sql.Close()

	// Create the table based off of the regex db map
	columns := ""
	for _, value := range s.RegexConfig.DBMapping {
		columns += value + " VARCHAR, "
	}

	columns = columns[:len(columns)-2] // Remove the trailing comma and space

	_, err = sql.Exec("CREATE TABLE IF NOT EXISTS logs ( ? )", columns)
	if err != nil {
		return err
	}

	return nil
}

// Write implements [Driver].
func (s *DuckDBWriter) Write(data buffer.BufferTransferData) error {
	// Connect to the DuckDB database and write the data.
	panic("not implemented")
}

var _ Driver = (*DuckDBWriter)(nil)
