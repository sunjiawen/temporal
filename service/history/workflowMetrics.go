// Copyright (c) 2019 Uber Technologies, Inc.
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
	"time"

	commonproto "go.temporal.io/temporal-proto/common"
	"go.temporal.io/temporal-proto/enums"

	"github.com/temporalio/temporal/common/metrics"
	"github.com/temporalio/temporal/common/persistence"
)

func emitWorkflowHistoryStats(
	metricsClient metrics.Client,
	namespace string,
	historySize int,
	historyCount int,
) {

	sizeScope := metricsClient.Scope(metrics.ExecutionSizeStatsScope, metrics.NamespaceTag(namespace))
	countScope := metricsClient.Scope(metrics.ExecutionCountStatsScope, metrics.NamespaceTag(namespace))

	sizeScope.RecordTimer(metrics.HistorySize, time.Duration(historySize))
	countScope.RecordTimer(metrics.HistoryCount, time.Duration(historyCount))
}

func emitWorkflowExecutionStats(
	metricsClient metrics.Client,
	namespace string,
	stats *persistence.MutableStateStats,
	executionInfoHistorySize int64,
) {

	if stats == nil {
		return
	}

	sizeScope := metricsClient.Scope(metrics.ExecutionSizeStatsScope, metrics.NamespaceTag(namespace))
	countScope := metricsClient.Scope(metrics.ExecutionCountStatsScope, metrics.NamespaceTag(namespace))

	sizeScope.RecordTimer(metrics.HistorySize, time.Duration(executionInfoHistorySize))
	sizeScope.RecordTimer(metrics.MutableStateSize, time.Duration(stats.MutableStateSize))
	sizeScope.RecordTimer(metrics.ExecutionInfoSize, time.Duration(stats.MutableStateSize))
	sizeScope.RecordTimer(metrics.ActivityInfoSize, time.Duration(stats.ActivityInfoSize))
	sizeScope.RecordTimer(metrics.TimerInfoSize, time.Duration(stats.TimerInfoSize))
	sizeScope.RecordTimer(metrics.ChildInfoSize, time.Duration(stats.ChildInfoSize))
	sizeScope.RecordTimer(metrics.SignalInfoSize, time.Duration(stats.SignalInfoSize))
	sizeScope.RecordTimer(metrics.BufferedEventsSize, time.Duration(stats.BufferedEventsSize))

	countScope.RecordTimer(metrics.ActivityInfoCount, time.Duration(stats.ActivityInfoCount))
	countScope.RecordTimer(metrics.TimerInfoCount, time.Duration(stats.TimerInfoCount))
	countScope.RecordTimer(metrics.ChildInfoCount, time.Duration(stats.ChildInfoCount))
	countScope.RecordTimer(metrics.SignalInfoCount, time.Duration(stats.SignalInfoCount))
	countScope.RecordTimer(metrics.RequestCancelInfoCount, time.Duration(stats.RequestCancelInfoCount))
	countScope.RecordTimer(metrics.BufferedEventsCount, time.Duration(stats.BufferedEventsCount))
}

func emitSessionUpdateStats(
	metricsClient metrics.Client,
	namespace string,
	stats *persistence.MutableStateUpdateSessionStats,
) {

	if stats == nil {
		return
	}

	sizeScope := metricsClient.Scope(metrics.SessionSizeStatsScope, metrics.NamespaceTag(namespace))
	countScope := metricsClient.Scope(metrics.SessionCountStatsScope, metrics.NamespaceTag(namespace))

	sizeScope.RecordTimer(metrics.MutableStateSize, time.Duration(stats.MutableStateSize))
	sizeScope.RecordTimer(metrics.ExecutionInfoSize, time.Duration(stats.ExecutionInfoSize))
	sizeScope.RecordTimer(metrics.ActivityInfoSize, time.Duration(stats.ActivityInfoSize))
	sizeScope.RecordTimer(metrics.TimerInfoSize, time.Duration(stats.TimerInfoSize))
	sizeScope.RecordTimer(metrics.ChildInfoSize, time.Duration(stats.ChildInfoSize))
	sizeScope.RecordTimer(metrics.SignalInfoSize, time.Duration(stats.SignalInfoSize))
	sizeScope.RecordTimer(metrics.BufferedEventsSize, time.Duration(stats.BufferedEventsSize))

	countScope.RecordTimer(metrics.ActivityInfoCount, time.Duration(stats.ActivityInfoCount))
	countScope.RecordTimer(metrics.TimerInfoCount, time.Duration(stats.TimerInfoCount))
	countScope.RecordTimer(metrics.ChildInfoCount, time.Duration(stats.ChildInfoCount))
	countScope.RecordTimer(metrics.SignalInfoCount, time.Duration(stats.SignalInfoCount))
	countScope.RecordTimer(metrics.RequestCancelInfoCount, time.Duration(stats.RequestCancelInfoCount))
	countScope.RecordTimer(metrics.DeleteActivityInfoCount, time.Duration(stats.DeleteActivityInfoCount))
	countScope.RecordTimer(metrics.DeleteTimerInfoCount, time.Duration(stats.DeleteTimerInfoCount))
	countScope.RecordTimer(metrics.DeleteChildInfoCount, time.Duration(stats.DeleteChildInfoCount))
	countScope.RecordTimer(metrics.DeleteSignalInfoCount, time.Duration(stats.DeleteSignalInfoCount))
	countScope.RecordTimer(metrics.DeleteRequestCancelInfoCount, time.Duration(stats.DeleteRequestCancelInfoCount))
}

func emitWorkflowCompletionStats(
	metricsClient metrics.Client,
	namespace string,
	taskList string,
	event *commonproto.HistoryEvent,
) {
	scope := metricsClient.Scope(
		metrics.WorkflowCompletionStatsScope,
		metrics.NamespaceTag(namespace),
		metrics.TaskListTag(taskList),
	)

	switch event.EventType {
	case enums.EventTypeWorkflowExecutionCompleted:
		scope.IncCounter(metrics.WorkflowSuccessCount)
	case enums.EventTypeWorkflowExecutionCanceled:
		scope.IncCounter(metrics.WorkflowCancelCount)
	case enums.EventTypeWorkflowExecutionFailed:
		scope.IncCounter(metrics.WorkflowFailedCount)
	case enums.EventTypeWorkflowExecutionTimedOut:
		scope.IncCounter(metrics.WorkflowTimeoutCount)
	case enums.EventTypeWorkflowExecutionTerminated:
		scope.IncCounter(metrics.WorkflowTerminateCount)
	}
}
