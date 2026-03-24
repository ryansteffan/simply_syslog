package buffer

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/ryansteffan/simply_syslog/pkg/applogger"
)

type WriteMode byte

const (
	WMF  WriteMode = 1
	WMDB WriteMode = 2
	WMS  WriteMode = 4
)

type ParsedSyslogData struct {
	RawMessage    []byte
	ParsedMessage map[string]string
}

type Buffer interface {
	Add(item ParsedSyslogData) error
	Flush() error
	StreamReader(wg *sync.WaitGroup)
	StreamWriter()
	WriteHandler() error
	MonitorAge(wg *sync.WaitGroup)
}

type SyslogWriteBuffer struct {
	Data           []ParsedSyslogData
	MaxLen         int
	MaxAge         time.Time
	MaxAgeDuration time.Duration
	CurrentLen     int
	CurrentAge     time.Time
	InChannel      chan ParsedSyslogData
	OutChannel     chan ParsedSyslogData
	Mode           WriteMode
	FilePath       *string
	Logger         applogger.Logger
}

func NewSyslogWriteBuffer(
	maxLen int,
	maxAge int,
	inChannel chan ParsedSyslogData,
	outChannel chan ParsedSyslogData,
	mode WriteMode,
	filePath *string,
	logger applogger.Logger,
) *SyslogWriteBuffer {
	maxAgeDuration := time.Duration(maxAge) * time.Second
	return &SyslogWriteBuffer{
		Data:           make([]ParsedSyslogData, 0, maxLen),
		MaxLen:         maxLen,
		CurrentLen:     0,
		MaxAgeDuration: maxAgeDuration,
		MaxAge:         time.Now().Add(maxAgeDuration),
		CurrentAge:     time.Now(),
		InChannel:      inChannel,
		OutChannel:     outChannel,
		Mode:           mode,
		FilePath:       filePath,
		Logger:         logger,
	}
}

// Add implements Buffer.
func (s *SyslogWriteBuffer) Add(item ParsedSyslogData) error {
	s.Logger.Debug("Adding item to buffer")
	if s.CurrentLen < s.MaxLen {
		s.Data = append(s.Data, item)
		s.CurrentLen++
		s.Logger.Debug("Item added. CurrentLen: " + fmt.Sprint(s.CurrentLen))
		return nil
	}
	s.Logger.Info("Buffer full, triggering WriteHandler")
	if err := s.WriteHandler(); err != nil {
		return err
	}
	s.Data = append(s.Data, item)
	s.CurrentLen++
	s.Logger.Debug("Item added after flush. CurrentLen: " + fmt.Sprint(s.CurrentLen))
	return nil
}

// Flush implements Buffer.
func (s *SyslogWriteBuffer) Flush() error {
	s.Logger.Info("Flushing buffer")
	s.Data = make([]ParsedSyslogData, 0, s.MaxLen)
	s.CurrentLen = 0
	s.MaxAge = time.Now().Add(s.MaxAgeDuration)
	s.CurrentAge = time.Now()
	s.Logger.Debug("Buffer flushed. CurrentLen reset to 0.")
	return nil
}

// WriteHandler implements Buffer.
func (s *SyslogWriteBuffer) WriteHandler() error {
	s.Logger.Info("WriteHandler triggered. Mode: " + fmt.Sprint(s.Mode))
	switch s.Mode {
	case WMF:
		s.Logger.Debug("Calling FileWriter")
		s.FileWriter()
	case WMDB:
		s.Logger.Debug("Calling DatabaseWriter")
		s.DatabaseWriter()
	case WMS:
		s.Logger.Debug("Calling StreamWriter")
		s.StreamWriter()
	case WMF + WMS:
		s.Logger.Debug("Calling FileWriter and StreamWriter")
		s.FileWriter()
		s.StreamWriter()
	case WMDB + WMS:
		s.Logger.Debug("Calling DatabaseWriter and StreamWriter")
		s.DatabaseWriter()
		s.StreamWriter()
	case WMF + WMDB:
		s.Logger.Debug("Calling FileWriter and DatabaseWriter")
		s.FileWriter()
		s.DatabaseWriter()
	case WMF + WMDB + WMS:
		s.Logger.Debug("Calling FileWriter, DatabaseWriter, and StreamWriter")
		s.FileWriter()
		s.DatabaseWriter()
		s.StreamWriter()
	default:
		s.Logger.Error("Invalid write mode")
		return errors.New("invalid write mode")
	}
	s.Flush()
	s.Logger.Info("WriteHandler completed and buffer flushed.")
	return nil
}

// MonitorAge implements Buffer.
func (s *SyslogWriteBuffer) MonitorAge(wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(s.MaxAgeDuration)
	defer ticker.Stop()
	for {
		<-ticker.C
		if s.CurrentLen > 0 && time.Since(s.CurrentAge) >= s.MaxAgeDuration {
			s.Logger.Info("MaxAge reached, triggering WriteHandler")
			s.WriteHandler()
		}
		ticker.Reset(s.MaxAgeDuration)
	}
}

// StreamReader implements Buffer.
func (s *SyslogWriteBuffer) StreamReader(wg *sync.WaitGroup) {
	defer wg.Done()

	for item := range s.InChannel {
		s.Add(item)
	}
}

// StreamWriter implements Buffer.
func (s *SyslogWriteBuffer) StreamWriter() {
	for _, item := range s.Data {
		s.OutChannel <- item
	}
}

func (s *SyslogWriteBuffer) FileWriter() error {
	if s.FilePath == nil {
		return errors.New("file path is nil")
	}

	file, err := os.OpenFile(*s.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer file.Close()

	var lines []byte
	for _, item := range s.Data {
		line := append(lines, item.RawMessage...)
		line = append(line, '\n')
		lines = line
	}
	_, err = file.Write(lines)
	if err != nil {
		return err
	}
	return nil
}

func (s *SyslogWriteBuffer) DatabaseWriter() error {
	panic("unimplemented")
}

var _ Buffer = (*SyslogWriteBuffer)(nil)
