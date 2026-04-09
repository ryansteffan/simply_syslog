package server

import "github.com/ryansteffan/simply_syslog/internal/pipeline"

func TCPServerProcessor(api pipeline.ProcessorAPI[string, ServerTransferData]) {}
