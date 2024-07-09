package pipeline

import (
	"context"
	"fmt"
	"github.com/marcodd23/go-micro-core/pkg/logmgr"
	"strconv"
	"sync"
)

type contextKey string

const (
	workerIDKey     contextKey = "workerID"
	pipelineNameKey contextKey = "pipelineName"
)

// Config holds the configuration for each pipeline, including its input and output channels, and the number of workers.
type Config struct {
	Pipeline   *Pipeline
	InputChan  chan Message
	OutputChan chan Message
	NumWorkers int
}

// Orchestrator orchestrates multiple pipelines, each with its own input and output channels and a configurable number of workers.
type Orchestrator struct {
	pipelines map[string]Config
}

// NewOrchestrator creates a new Orchestrator.
//
// Parameters:
// - pipelines: A map where each key is a pipeline identifier and the value is the Config.
//
// Returns:
// - A new instance of Orchestrator.
func NewOrchestrator(pipelines map[string]Config) *Orchestrator {
	return &Orchestrator{
		pipelines: pipelines,
	}
}

// Execute runs all pipelines, each with its configured number of workers.
//
// This method starts multiple goroutines for each pipeline based on its configuration.
// Each goroutine listens for messages on its input channel, processes them through
// the pipeline, and sends the results to the output channel.
//
// Parameters:
// - cancelCtx: A context that can be used to cancel the processing.
// - wg: A wait group to wait for all processing to complete.
func (o *Orchestrator) Execute(cancelCtx context.Context, wg *sync.WaitGroup) {
	for name, config := range o.pipelines {
		for j := 0; j < config.NumWorkers; j++ {
			wg.Add(1)
			workerID := j
			go func(pipelineName string, pipelineConfig Config, workerID int) {
				defer wg.Done()
				o.processMessages(cancelCtx, pipelineName, pipelineConfig, workerID)
			}(name, config, workerID)
		}
	}
}

// processMessages is the main loop for processing messages.
//
// This method is executed by each worker goroutine. It listens for messages on the
// input channel, processes each message through the pipeline, and sends the results
// to the output channel. If the context is cancelled, the worker stops processing.
//
// Parameters:
// - ctx: The context used for cancellation.
// - pipelineID: The ID of the pipeline being processed.
// - pipelineConfig: The configuration of the pipeline being processed.
// - workerID: The ID of the worker processing the messages.
func (o *Orchestrator) processMessages(ctx context.Context, pipelineName string, pipelineConfig Config, workerID int) {
	for {
		select {
		case <-ctx.Done():
			// Context was cancelled, stop processing
			logmgr.GetLogger().LogInfo(context.Background(), fmt.Sprintf("Pipeline %s Worker %d stopping due to context cancellation", pipelineName, workerID))
			return
		case msg, ok := <-pipelineConfig.InputChan:
			if !ok {
				// Channel was closed, stop processing
				return
			}

			// Add worker ID and pipeline name to the context
			msgCtx := context.WithValue(ctx, workerIDKey, strconv.Itoa(workerID))
			msgCtx = context.WithValue(msgCtx, pipelineNameKey, pipelineName)

			// Process the message
			processedMsg, err := pipelineConfig.Pipeline.Process(msgCtx, msg)
			if err != nil {
				logmgr.GetLogger().LogError(msgCtx, fmt.Sprintf("Pipeline: %s, Worker: %d, error processing message %s: %v", pipelineName, workerID, msg.GetMsgRefId(), err))
				continue
			}
			logmgr.GetLogger().LogInfo(msgCtx, fmt.Sprintf("Pipeline: %s, Worker: %d, completed processing message %s successfully", pipelineName, workerID, processedMsg.GetMsgRefId()))

			// Send to the output channel if provided
			if pipelineConfig.OutputChan != nil {
				pipelineConfig.OutputChan <- processedMsg
			}
		}
	}
}
