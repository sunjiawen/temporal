// Copyright (c) 2020 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package history

import (
	"context"
	"time"

	commonproto "go.temporal.io/temporal-proto/common"
	"go.temporal.io/temporal-proto/enums"
	"go.temporal.io/temporal-proto/serviceerror"

	m "github.com/temporalio/temporal/.gen/proto/matchingservice"

	"github.com/temporalio/temporal/.gen/proto/persistenceblobs"
	"github.com/temporalio/temporal/client/matching"
	"github.com/temporalio/temporal/common"
	"github.com/temporalio/temporal/common/log"
	"github.com/temporalio/temporal/common/log/tag"
	"github.com/temporalio/temporal/common/metrics"
	"github.com/temporalio/temporal/common/persistence"
	"github.com/temporalio/temporal/common/primitives"
	"github.com/temporalio/temporal/service/worker/archiver"
)

const (
	transferActiveTaskDefaultTimeout = 30 * time.Second
)

type (
	transferQueueTaskExecutorBase struct {
		shard          ShardContext
		historyService *historyEngineImpl
		cache          *historyCache
		logger         log.Logger
		metricsClient  metrics.Client
		matchingClient matching.Client
		visibilityMgr  persistence.VisibilityManager
		config         *Config
	}
)

func newTransferQueueTaskExecutorBase(
	shard ShardContext,
	historyService *historyEngineImpl,
	logger log.Logger,
	metricsClient metrics.Client,
	config *Config,
) *transferQueueTaskExecutorBase {
	return &transferQueueTaskExecutorBase{
		shard:          shard,
		historyService: historyService,
		cache:          historyService.historyCache,
		logger:         logger,
		metricsClient:  metricsClient,
		matchingClient: shard.GetService().GetMatchingClient(),
		visibilityMgr:  shard.GetService().GetVisibilityManager(),
		config:         config,
	}
}

func (t *transferQueueTaskExecutorBase) getNamespaceIDAndWorkflowExecution(
	task *persistenceblobs.TransferTaskInfo,
) (string, commonproto.WorkflowExecution) {

	return primitives.UUIDString(task.NamespaceID), commonproto.WorkflowExecution{
		WorkflowId: task.WorkflowID,
		RunId:      primitives.UUIDString(task.RunID),
	}
}

func (t *transferQueueTaskExecutorBase) pushActivity(
	task *persistenceblobs.TransferTaskInfo,
	activityScheduleToStartTimeout int32,
) error {

	ctx, cancel := context.WithTimeout(context.Background(), transferActiveTaskDefaultTimeout)
	defer cancel()

	if task.TaskType != persistence.TransferTaskTypeActivityTask {
		t.logger.Fatal("Cannot process non activity task", tag.TaskType(task.GetTaskType()))
	}

	_, err := t.matchingClient.AddActivityTask(ctx, &m.AddActivityTaskRequest{
		NamespaceUUID:       primitives.UUIDString(task.TargetNamespaceID),
		SourceNamespaceUUID: primitives.UUIDString(task.NamespaceID),
		Execution: &commonproto.WorkflowExecution{
			WorkflowId: task.WorkflowID,
			RunId:      primitives.UUIDString(task.RunID),
		},
		TaskList:                      &commonproto.TaskList{Name: task.TaskList},
		ScheduleId:                    task.ScheduleID,
		ScheduleToStartTimeoutSeconds: activityScheduleToStartTimeout,
	})

	return err
}

func (t *transferQueueTaskExecutorBase) pushDecision(
	task *persistenceblobs.TransferTaskInfo,
	tasklist *commonproto.TaskList,
	decisionScheduleToStartTimeout int32,
) error {

	ctx, cancel := context.WithTimeout(context.Background(), transferActiveTaskDefaultTimeout)
	defer cancel()

	if task.TaskType != persistence.TransferTaskTypeDecisionTask {
		t.logger.Fatal("Cannot process non decision task", tag.TaskType(task.GetTaskType()))
	}

	_, err := t.matchingClient.AddDecisionTask(ctx, &m.AddDecisionTaskRequest{
		NamespaceUUID: primitives.UUIDString(task.NamespaceID),
		Execution: &commonproto.WorkflowExecution{
			WorkflowId: task.WorkflowID,
			RunId:      primitives.UUIDString(task.RunID),
		},
		TaskList:                      tasklist,
		ScheduleId:                    task.ScheduleID,
		ScheduleToStartTimeoutSeconds: decisionScheduleToStartTimeout,
	})
	return err
}

func (t *transferQueueTaskExecutorBase) recordWorkflowStarted(
	namespaceID string,
	workflowID string,
	runID string,
	workflowTypeName string,
	startTimeUnixNano int64,
	executionTimeUnixNano int64,
	workflowTimeout int32,
	taskID int64,
	visibilityMemo *commonproto.Memo,
	searchAttributes map[string][]byte,
) error {

	namespace := defaultNamespace

	if namespaceEntry, err := t.shard.GetNamespaceCache().GetNamespaceByID(namespaceID); err != nil {
		if _, ok := err.(*serviceerror.NotFound); !ok {
			return err
		}
	} else {
		namespace = namespaceEntry.GetInfo().Name
		// if sampled for longer retention is enabled, only record those sampled events
		if namespaceEntry.IsSampledForLongerRetentionEnabled(workflowID) &&
			!namespaceEntry.IsSampledForLongerRetention(workflowID) {
			return nil
		}
	}

	request := &persistence.RecordWorkflowExecutionStartedRequest{
		NamespaceUUID: namespaceID,
		Namespace:     namespace,
		Execution: commonproto.WorkflowExecution{
			WorkflowId: workflowID,
			RunId:      runID,
		},
		WorkflowTypeName:   workflowTypeName,
		StartTimestamp:     startTimeUnixNano,
		ExecutionTimestamp: executionTimeUnixNano,
		WorkflowTimeout:    int64(workflowTimeout),
		TaskID:             taskID,
		Memo:               visibilityMemo,
		SearchAttributes:   searchAttributes,
	}

	return t.visibilityMgr.RecordWorkflowExecutionStarted(request)
}

func (t *transferQueueTaskExecutorBase) upsertWorkflowExecution(
	namespaceID string,
	workflowID string,
	runID string,
	workflowTypeName string,
	startTimeUnixNano int64,
	executionTimeUnixNano int64,
	workflowTimeout int32,
	taskID int64,
	visibilityMemo *commonproto.Memo,
	searchAttributes map[string][]byte,
) error {

	namespace := defaultNamespace
	namespaceEntry, err := t.shard.GetNamespaceCache().GetNamespaceByID(namespaceID)
	if err != nil {
		if _, ok := err.(*serviceerror.NotFound); !ok {
			return err
		}
	} else {
		namespace = namespaceEntry.GetInfo().Name
	}

	request := &persistence.UpsertWorkflowExecutionRequest{
		NamespaceUUID: namespaceID,
		Namespace:     namespace,
		Execution: commonproto.WorkflowExecution{
			WorkflowId: workflowID,
			RunId:      runID,
		},
		WorkflowTypeName:   workflowTypeName,
		StartTimestamp:     startTimeUnixNano,
		ExecutionTimestamp: executionTimeUnixNano,
		WorkflowTimeout:    int64(workflowTimeout),
		TaskID:             taskID,
		Memo:               visibilityMemo,
		SearchAttributes:   searchAttributes,
	}

	return t.visibilityMgr.UpsertWorkflowExecution(request)
}

func (t *transferQueueTaskExecutorBase) recordWorkflowClosed(
	namespaceID string,
	workflowID string,
	runID string,
	workflowTypeName string,
	startTimeUnixNano int64,
	executionTimeUnixNano int64,
	endTimeUnixNano int64,
	closeStatus enums.WorkflowExecutionCloseStatus,
	historyLength int64,
	taskID int64,
	visibilityMemo *commonproto.Memo,
	searchAttributes map[string][]byte,
) error {

	// Record closing in visibility store
	retentionSeconds := int64(0)
	namespace := defaultNamespace
	recordWorkflowClose := true
	archiveVisibility := false

	namespaceEntry, err := t.shard.GetNamespaceCache().GetNamespaceByID(namespaceID)
	if err != nil && !isWorkflowNotExistError(err) {
		return err
	}

	if err == nil {
		// retention in namespace config is in days, convert to seconds
		retentionSeconds = int64(namespaceEntry.GetRetentionDays(workflowID)) * int64(secondsInDay)
		namespace = namespaceEntry.GetInfo().Name
		// if sampled for longer retention is enabled, only record those sampled events
		if namespaceEntry.IsSampledForLongerRetentionEnabled(workflowID) &&
			!namespaceEntry.IsSampledForLongerRetention(workflowID) {
			recordWorkflowClose = false
		}

		clusterConfiguredForVisibilityArchival := t.shard.GetService().GetArchivalMetadata().GetVisibilityConfig().ClusterConfiguredForArchival()
		namespaceConfiguredForVisibilityArchival := namespaceEntry.GetConfig().VisibilityArchivalStatus == enums.ArchivalStatusEnabled
		archiveVisibility = clusterConfiguredForVisibilityArchival && namespaceConfiguredForVisibilityArchival
	}

	if recordWorkflowClose {
		if err := t.visibilityMgr.RecordWorkflowExecutionClosed(&persistence.RecordWorkflowExecutionClosedRequest{
			NamespaceUUID: namespaceID,
			Namespace:     namespace,
			Execution: commonproto.WorkflowExecution{
				WorkflowId: workflowID,
				RunId:      runID,
			},
			WorkflowTypeName:   workflowTypeName,
			StartTimestamp:     startTimeUnixNano,
			ExecutionTimestamp: executionTimeUnixNano,
			CloseTimestamp:     endTimeUnixNano,
			Status:             closeStatus,
			HistoryLength:      historyLength,
			RetentionSeconds:   retentionSeconds,
			TaskID:             taskID,
			Memo:               visibilityMemo,
			SearchAttributes:   searchAttributes,
		}); err != nil {
			return err
		}
	}

	if archiveVisibility {
		ctx, cancel := context.WithTimeout(context.Background(), t.config.TransferProcessorVisibilityArchivalTimeLimit())
		defer cancel()
		_, err := t.historyService.archivalClient.Archive(ctx, &archiver.ClientRequest{
			ArchiveRequest: &archiver.ArchiveRequest{
				NamespaceID:        namespaceID,
				Namespace:          namespace,
				WorkflowID:         workflowID,
				RunID:              runID,
				WorkflowTypeName:   workflowTypeName,
				StartTimestamp:     startTimeUnixNano,
				ExecutionTimestamp: executionTimeUnixNano,
				CloseTimestamp:     endTimeUnixNano,
				CloseStatus:        closeStatus,
				HistoryLength:      historyLength,
				Memo:               visibilityMemo,
				SearchAttributes:   searchAttributes,
				VisibilityURI:      namespaceEntry.GetConfig().VisibilityArchivalURI,
				URI:                namespaceEntry.GetConfig().HistoryArchivalURI,
				Targets:            []archiver.ArchivalTarget{archiver.ArchiveTargetVisibility},
			},
			CallerService:        common.HistoryServiceName,
			AttemptArchiveInline: true, // archive visibility inline by default
		})
		return err
	}
	return nil
}

// Argument startEvent is to save additional call of msBuilder.GetStartEvent
func getWorkflowExecutionTimestamp(
	msBuilder mutableState,
	startEvent *commonproto.HistoryEvent,
) time.Time {
	// Use value 0 to represent workflows that don't need backoff. Since ES doesn't support
	// comparison between two field, we need a value to differentiate them from cron workflows
	// or later runs of a workflow that needs retry.
	executionTimestamp := time.Unix(0, 0)
	if startEvent == nil {
		return executionTimestamp
	}

	if backoffSeconds := startEvent.GetWorkflowExecutionStartedEventAttributes().GetFirstDecisionTaskBackoffSeconds(); backoffSeconds != 0 {
		startTimestamp := time.Unix(0, startEvent.GetTimestamp())
		executionTimestamp = startTimestamp.Add(time.Duration(backoffSeconds) * time.Second)
	}
	return executionTimestamp
}

func getWorkflowMemo(
	memo map[string][]byte,
) *commonproto.Memo {

	if memo == nil {
		return nil
	}
	return &commonproto.Memo{Fields: memo}
}

func copySearchAttributes(
	input map[string][]byte,
) map[string][]byte {

	if input == nil {
		return nil
	}

	result := make(map[string][]byte)
	for k, v := range input {
		val := make([]byte, len(v))
		copy(val, v)
		result[k] = val
	}
	return result
}

func isWorkflowNotExistError(err error) bool {
	_, ok := err.(*serviceerror.NotFound)
	return ok
}
