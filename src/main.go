package main

import (
	"github.com/geometry-labs/icon-go-etl/config"
	"github.com/geometry-labs/icon-go-etl/extractor"
	"github.com/geometry-labs/icon-go-etl/global"
	"github.com/geometry-labs/icon-go-etl/logging"
	"github.com/geometry-labs/icon-go-etl/transformer"
)

func main() {
	config.ReadEnvironment()

	logging.Init()

	extractor.StartManager()

	transformer.StartTransformer()

	global.WaitShutdownSig()
}