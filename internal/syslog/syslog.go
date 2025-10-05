// Defines the struct for syslog parsers in the server,
// as well as a default implementation for it.
package syslog

import (
	"encoding/json"
	"errors"
	"os"
	"regexp"
	"sync"

	"github.com/ryansteffan/simply_syslog/pkg/applogger"
)

type Format struct {
	Version int    `json:"version"`
	Name    string `json:"name"`
	Format  string `json:"format"`
}

type CompiledFormat struct {
	Name   string
	Format *regexp.Regexp
}

type EvenDrivenSyslogParser struct {
	Formats         *[]Format
	CompiledFormats *[]CompiledFormat
	MessageChannel  chan []byte
}

// DetectFormat implements SyslogParser.
func (d *EvenDrivenSyslogParser) DetectFormat(message []byte) (*CompiledFormat, error) {
	if d.CompiledFormats == nil {
		return nil, errors.New("no compiled formats available to detect")
	}
	for _, compiledFormat := range *d.CompiledFormats {
		// Require the message to match the entire format regex exactly
		if compiledFormat.Format.Match(message) {
			matched := compiledFormat.Format.Find(message)
			if matched != nil && string(matched) == string(message) {
				return &compiledFormat, nil
			}
		}
	}
	return nil, errors.New("no matching format found")
}

func NewEvenDrivenSyslogParser(path string) (*EvenDrivenSyslogParser, error) {

	parser := &EvenDrivenSyslogParser{
		Formats:         nil,
		CompiledFormats: nil,
		MessageChannel:  make(chan []byte),
	}

	switch path {
	case "ENV":
		err := parser.LoadFormatsFromENV()
		if err != nil {
			return nil, err
		}
	default:
		err := parser.LoadFormatsFromFile(path)
		if err != nil {
			return nil, err
		}
	}

	return parser, nil
}

// LoadFormatsFromDB implements SyslogParser.
func (d *EvenDrivenSyslogParser) LoadFormatsFromDB() error {
	panic("unimplemented")
}

// LoadFormatsFromENV implements SyslogParser.
func (d *EvenDrivenSyslogParser) LoadFormatsFromENV() error {
	envFormats := os.Getenv("SYSLOG_FORMATS")
	if envFormats != "" {
		err := json.Unmarshal([]byte(envFormats), &d.Formats)
		if err != nil {
			return err
		}

		err = d.compileFormats()
		if err != nil {
			return err
		}

		return nil
	}
	return errors.New("SYSLOG_FORMATS from ENV not set")
}

// LoadFormatsFromFile implements SyslogParser.
func (d *EvenDrivenSyslogParser) LoadFormatsFromFile(path string) error {
	fileContent, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if len(fileContent) == 0 {
		return errors.New("regex file is empty")
	}

	var formats []Format

	err = json.Unmarshal(fileContent, &formats)
	if err != nil {
		return err
	}

	if len(formats) == 0 {
		return errors.New("no formats found in regex file")
	}
	d.Formats = &formats

	err = d.compileFormats()
	if err != nil {
		return err
	}

	return nil
}

// ParseMessage implements SyslogParser.
func (d *EvenDrivenSyslogParser) ParseMessage(message []byte, format *CompiledFormat) (map[string]string, error) {
	if format == nil {
		return nil, errors.New("no format provided for parsing")
	}

	itemMap := make(map[string]string)
	data := format.Format.FindSubmatch(message)
	names := format.Format.SubexpNames()

	for i, name := range names {
		if i > 0 && i < len(data) {
			itemMap[name] = string(data[i])
		}
	}
	return itemMap, nil
}

func (d *EvenDrivenSyslogParser) compileFormats() error {
	if d.Formats == nil {
		return errors.New("no formats loaded to compile")
	}
	compiledFormats := make([]CompiledFormat, 0, len(*d.Formats))
	for _, format := range *d.Formats {
		compiled, err := regexp.Compile(format.Format)
		if err != nil {
			return err
		}
		compiledFormat := CompiledFormat{
			Name:   format.Name,
			Format: compiled,
		}
		compiledFormats = append(compiledFormats, compiledFormat)
	}
	d.CompiledFormats = &compiledFormats
	return nil
}

type SyslogParser interface {
	LoadFormatsFromENV() error
	LoadFormatsFromFile(path string) error
	LoadFormatsFromDB() error
	ParseMessage(message []byte, format *CompiledFormat) (map[string]string, error)
	DetectFormat(message []byte) (*CompiledFormat, error)
}

var _ SyslogParser = (*EvenDrivenSyslogParser)(nil)

func HandleSyslogMessages(
	parser SyslogParser,
	messageChannel chan []byte,
	writeBufferInChannel chan map[string]string,
	logger applogger.Logger,
	waitGroup *sync.WaitGroup,
) {
	defer waitGroup.Done()

	for message := range messageChannel {
		format, err := parser.DetectFormat(message)
		if err != nil {
			logger.Warn("No matching format found for message: " + string(message))
			continue
		}

		parsedMessage, err := parser.ParseMessage(message, format)
		if err != nil {
			logger.Warn("Error parsing message: " + err.Error())
			continue
		}

		writeBufferInChannel <- parsedMessage
	}
}
