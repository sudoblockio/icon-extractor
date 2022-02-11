package extractor

import (
	"github.com/geometry-labs/icon-go-etl/transformer"
	"go.uber.org/zap"
)

func StartManager() {

	go startManager(1)
}

func startManager(blockNumber int64) {

	extractorQueueChannel := make(chan int64)
	extractorCommitChannel := make(chan int64)

	extractor := Extractor{
		blockNumberQueue:  extractorQueueChannel,
		blockNumberCommit: extractorCommitChannel,
		blockOutput:       transformer.RawBlockChannel,
	}
	extractor.Start()

	for {
		extractorQueueChannel <- blockNumber

		commitBlockNumber := <-extractorCommitChannel
		zap.S().Info("COMMIT BLOCK #", commitBlockNumber)

		blockNumber++
	}
}
