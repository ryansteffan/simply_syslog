package drivers

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/duckdb/duckdb-go/v2"
	_ "github.com/duckdb/duckdb-go/v2"
	"github.com/ryansteffan/simply_syslog/internal/buffer"
	"github.com/ryansteffan/simply_syslog/internal/config"
	"github.com/ryansteffan/simply_syslog/pkg/applogger"
)

type DuckDBWriter struct {
	Name        string
	Enabled     bool
	RegexConfig *config.RegexConfig
	Logger      applogger.Logger
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

	if logLevel, ok := conf.Options["log_level"]; ok {
		logLevelInt, err := strconv.Atoi(logLevel)
		if err != nil {
			return errors.New("invalid log_level directive in writer config")
		}
		logger, err := applogger.NewLogger("DuckDBWriter", applogger.LogLevel(logLevelInt), applogger.CONSOLE, nil)
		if err != nil {
			return err
		}
		s.Logger = logger
	}

	// Pull the table map for the database
	regexConf, err := config.GetRegexConfig()
	if err != nil {
		return err
	}
	s.RegexConfig = regexConf

	// Create the path for the database
	if err := os.MkdirAll(filepath.Dir(s.Options.Path), 0755); err != nil {
		return err
	}

	// Create the database and required tables
	sql, err := sql.Open("duckdb", s.Options.Path)
	if err != nil {
		return err
	}
	defer sql.Close()

	// Create a sequence for the pk
	sqlStatement := `CREATE SEQUENCE IF NOT EXISTS syslog_id_seq START 1;`
	_, err = sql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	sqlStatement = `
	CREATE TABLE IF NOT EXISTS syslog (
		id INTEGER PRIMARY KEY DEFAULT nextval('syslog_id_seq'),
		priority INTEGER,
		timestamp TIMESTAMP,
		hostname VARCHAR,
		message VARCHAR,
		version VARCHAR,
		app_name VARCHAR,
		proc_id VARCHAR,
		msg_id VARCHAR,
		metadata JSON
	);
	`

	_, err = sql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

// Write implements [Driver].
func (s *DuckDBWriter) Write(data buffer.BufferTransferData) error {
	// Connect to the DuckDB database and write the data
	connector, err := duckdb.NewConnector(s.Options.Path, nil)
	if err != nil {
		return err
	}
	defer connector.Close()

	connection, err := connector.Connect(context.Background())
	if err != nil {
		return err
	}
	defer connection.Close()

	appender, err := duckdb.NewAppenderWithColumns(connection, "", "", "syslog", []string{
		"priority",
		"timestamp",
		"hostname",
		"message",
		"version",
		"app_name",
		"proc_id",
		"msg_id",
		"metadata",
	})
	if err != nil {
		return err
	}
	defer appender.Close()

	for _, item := range data.ParsedData {
		err := appendRow(item, appender)
		if err != nil {
			return err
		}
	}

	err = appender.Flush()
	if err != nil {
		return err
	}

	return nil
}

func appendRow(row map[string]string, appender *duckdb.Appender) error {
	var priority int
	if priStr, ok := row["priority"]; ok {
		priInt, err := strconv.Atoi(priStr)
		if err != nil {
			return errors.New("invalid priority value in parsed data")
		}
		priority = priInt
	} else {
		return errors.New("priority field not found in parsed data")
	}

	var timestamp time.Time
	if ts, ok := row["timestamp"]; ok {
		var parsedTime time.Time
		var parserErr error
		timeFormats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"Jan _2 15:04:05",      // RFC3164 format without the year
			"Jan _2 2006 15:04:05", // RFC3164 format with the year
		}

		errCount := 0
		for _, format := range timeFormats {
			parsedTime, parserErr = time.ParseInLocation(format, ts, time.Local)
			if parserErr == nil {
				break
			} else {
				errCount++
			}
		}

		if errCount == len(timeFormats) {
			return errors.New("invalid timestamp format in parsed data")
		}

		timestamp = parsedTime
	} else {
		return errors.New("timestamp field not found in parsed data")
	}

	var hostname string
	if hn, ok := row["hostname"]; ok {
		hostname = hn
	} else {
		return errors.New("hostname field not found in parsed data")
	}

	var message string
	if msg, ok := row["message"]; ok {
		message = msg
	} else {
		return errors.New("message field not found in parsed data")
	}

	var version string
	if ver, ok := row["version"]; ok {
		version = ver
	} else {
		version = ""
	}

	var appName string
	if app, ok := row["app_name"]; ok {
		appName = app
	} else {
		appName = ""
	}

	var procID string
	if proc, ok := row["proc_id"]; ok {
		procID = proc
	} else {
		procID = ""
	}

	var msgID string
	if msgid, ok := row["msg_id"]; ok {
		msgID = msgid
	} else {
		msgID = ""
	}

	err := appender.AppendRow(priority, timestamp, hostname, message, version, appName, procID, msgID, nil)
	if err != nil {
		return err
	}

	return nil
}

var _ Driver = (*DuckDBWriter)(nil)
