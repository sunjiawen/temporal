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

//go:generate mockgen -copyright_file ../../LICENSE -package $GOPACKAGE -source $GOFILE -destination replicationTaskHandler_mock.go

package namespace

import (
	commonproto "go.temporal.io/temporal-proto/common"
	"go.temporal.io/temporal-proto/enums"
	"go.temporal.io/temporal-proto/serviceerror"

	"github.com/temporalio/temporal/.gen/proto/replication"
	"github.com/temporalio/temporal/common/log"
	"github.com/temporalio/temporal/common/persistence"
)

var (
	// ErrEmptyNamespaceReplicationTask is the error to indicate empty replication task
	ErrEmptyNamespaceReplicationTask = serviceerror.NewInvalidArgument("empty namespace replication task")
	// ErrInvalidNamespaceOperation is the error to indicate empty namespace operation attribute
	ErrInvalidNamespaceOperation = serviceerror.NewInvalidArgument("invalid namespace operation attribute")
	// ErrInvalidNamespaceID is the error to indicate empty rID attribute
	ErrInvalidNamespaceID = serviceerror.NewInvalidArgument("invalid namespace ID attribute")
	// ErrInvalidNamespaceInfo is the error to indicate empty info attribute
	ErrInvalidNamespaceInfo = serviceerror.NewInvalidArgument("invalid namespace info attribute")
	// ErrInvalidNamespaceConfig is the error to indicate empty config attribute
	ErrInvalidNamespaceConfig = serviceerror.NewInvalidArgument("invalid namespace config attribute")
	// ErrInvalidNamespaceReplicationConfig is the error to indicate empty replication config attribute
	ErrInvalidNamespaceReplicationConfig = serviceerror.NewInvalidArgument("invalid namespace replication config attribute")
	// ErrInvalidNamespaceConfigVersion is the error to indicate empty config version attribute
	ErrInvalidNamespaceConfigVersion = serviceerror.NewInvalidArgument("invalid namespace config version attribute")
	// ErrInvalidNamespaceFailoverVersion is the error to indicate empty failover version attribute
	ErrInvalidNamespaceFailoverVersion = serviceerror.NewInvalidArgument("invalid namespace failover version attribute")
	// ErrInvalidNamespaceStatus is the error to indicate invalid namespace status
	ErrInvalidNamespaceStatus = serviceerror.NewInvalidArgument("invalid namespace status attribute")
	// ErrNameUUIDCollision is the error to indicate namespace name / UUID collision
	ErrNameUUIDCollision = serviceerror.NewInvalidArgument("namespace replication encounter name / UUID collision")
)

// NOTE: the counterpart of namespace replication transmission logic is in service/fropntend package

type (
	// ReplicationTaskExecutor is the interface which is to execute namespace replication task
	ReplicationTaskExecutor interface {
		Execute(task *replication.NamespaceTaskAttributes) error
	}

	namespaceReplicationTaskExecutorImpl struct {
		metadataManagerV2 persistence.MetadataManager
		logger            log.Logger
	}
)

// NewReplicationTaskExecutor create a new instance of namespace replicator
func NewReplicationTaskExecutor(
	metadataManagerV2 persistence.MetadataManager,
	logger log.Logger,
) ReplicationTaskExecutor {

	return &namespaceReplicationTaskExecutorImpl{
		metadataManagerV2: metadataManagerV2,
		logger:            logger,
	}
}

// Execute handles receiving of the namespace replication task
func (h *namespaceReplicationTaskExecutorImpl) Execute(task *replication.NamespaceTaskAttributes) error {
	if err := h.validateNamespaceReplicationTask(task); err != nil {
		return err
	}

	switch task.GetNamespaceOperation() {
	case enums.NamespaceOperationCreate:
		return h.handleNamespaceCreationReplicationTask(task)
	case enums.NamespaceOperationUpdate:
		return h.handleNamespaceUpdateReplicationTask(task)
	default:
		return ErrInvalidNamespaceOperation
	}
}

// handleNamespaceCreationReplicationTask handles the namespace creation replication task
func (h *namespaceReplicationTaskExecutorImpl) handleNamespaceCreationReplicationTask(task *replication.NamespaceTaskAttributes) error {
	// task already validated
	status, err := h.convertNamespaceStatusFromProto(task.Info.Status)
	if err != nil {
		return err
	}

	request := &persistence.CreateNamespaceRequest{
		Info: &persistence.NamespaceInfo{
			ID:          task.GetId(),
			Name:        task.Info.GetName(),
			Status:      status,
			Description: task.Info.GetDescription(),
			OwnerEmail:  task.Info.GetOwnerEmail(),
			Data:        task.Info.Data,
		},
		Config: &persistence.NamespaceConfig{
			Retention:                task.Config.GetWorkflowExecutionRetentionPeriodInDays(),
			EmitMetric:               task.Config.GetEmitMetric().GetValue(),
			HistoryArchivalStatus:    task.Config.GetHistoryArchivalStatus(),
			HistoryArchivalURI:       task.Config.GetHistoryArchivalURI(),
			VisibilityArchivalStatus: task.Config.GetVisibilityArchivalStatus(),
			VisibilityArchivalURI:    task.Config.GetVisibilityArchivalURI(),
		},
		ReplicationConfig: &persistence.NamespaceReplicationConfig{
			ActiveClusterName: task.ReplicationConfig.GetActiveClusterName(),
			Clusters:          h.convertClusterReplicationConfigFromProto(task.ReplicationConfig.Clusters),
		},
		IsGlobalNamespace: true, // local namespace will not be replicated
		ConfigVersion:     task.GetConfigVersion(),
		FailoverVersion:   task.GetFailoverVersion(),
	}

	_, err = h.metadataManagerV2.CreateNamespace(request)
	if err != nil {
		// SQL and Cassandra handle namespace UUID collision differently
		// here, whenever seeing a error replicating a namespace
		// do a check if there is a name / UUID collision

		recordExists := true
		resp, getErr := h.metadataManagerV2.GetNamespace(&persistence.GetNamespaceRequest{
			Name: task.Info.GetName(),
		})
		switch getErr.(type) {
		case nil:
			if resp.Info.ID != task.GetId() {
				return ErrNameUUIDCollision
			}
		case *serviceerror.NotFound:
			// no check is necessary
			recordExists = false
		default:
			// return the original err
			return err
		}

		resp, getErr = h.metadataManagerV2.GetNamespace(&persistence.GetNamespaceRequest{
			ID: task.GetId(),
		})
		switch getErr.(type) {
		case nil:
			if resp.Info.Name != task.Info.GetName() {
				return ErrNameUUIDCollision
			}
		case *serviceerror.NotFound:
			// no check is necessary
			recordExists = false
		default:
			// return the original err
			return err
		}

		if recordExists {
			// name -> id & id -> name check pass, this is duplication request
			return nil
		}
		return err
	}

	return err
}

// handleNamespaceUpdateReplicationTask handles the namespace update replication task
func (h *namespaceReplicationTaskExecutorImpl) handleNamespaceUpdateReplicationTask(task *replication.NamespaceTaskAttributes) error {
	// task already validated
	status, err := h.convertNamespaceStatusFromProto(task.Info.Status)
	if err != nil {
		return err
	}

	// first we need to get the current notification version since we need to it for conditional update
	metadata, err := h.metadataManagerV2.GetMetadata()
	if err != nil {
		return err
	}
	notificationVersion := metadata.NotificationVersion

	// plus, we need to check whether the config version is <= the config version set in the input
	// plus, we need to check whether the failover version is <= the failover version set in the input
	resp, err := h.metadataManagerV2.GetNamespace(&persistence.GetNamespaceRequest{
		Name: task.Info.GetName(),
	})
	if err != nil {
		if _, ok := err.(*serviceerror.NotFound); ok {
			// this can happen if the create namespace replication task is to processed.
			// e.g. new cluster which does not have anything
			return h.handleNamespaceCreationReplicationTask(task)
		}
		return err
	}

	recordUpdated := false
	request := &persistence.UpdateNamespaceRequest{
		Info:                        resp.Info,
		Config:                      resp.Config,
		ReplicationConfig:           resp.ReplicationConfig,
		ConfigVersion:               resp.ConfigVersion,
		FailoverVersion:             resp.FailoverVersion,
		FailoverNotificationVersion: resp.FailoverNotificationVersion,
		NotificationVersion:         notificationVersion,
	}

	if resp.ConfigVersion < task.GetConfigVersion() {
		recordUpdated = true
		request.Info = &persistence.NamespaceInfo{
			ID:          task.GetId(),
			Name:        task.Info.GetName(),
			Status:      status,
			Description: task.Info.GetDescription(),
			OwnerEmail:  task.Info.GetOwnerEmail(),
			Data:        task.Info.Data,
		}
		request.Config = &persistence.NamespaceConfig{
			Retention:                task.Config.GetWorkflowExecutionRetentionPeriodInDays(),
			EmitMetric:               task.Config.GetEmitMetric().GetValue(),
			HistoryArchivalStatus:    task.Config.GetHistoryArchivalStatus(),
			HistoryArchivalURI:       task.Config.GetHistoryArchivalURI(),
			VisibilityArchivalStatus: task.Config.GetVisibilityArchivalStatus(),
			VisibilityArchivalURI:    task.Config.GetVisibilityArchivalURI(),
		}
		if task.Config.GetBadBinaries() != nil {
			request.Config.BadBinaries = *task.Config.GetBadBinaries()
		}
		request.ReplicationConfig.Clusters = h.convertClusterReplicationConfigFromProto(task.ReplicationConfig.Clusters)
		request.ConfigVersion = task.GetConfigVersion()
	}
	if resp.FailoverVersion < task.GetFailoverVersion() {
		recordUpdated = true
		request.ReplicationConfig.ActiveClusterName = task.ReplicationConfig.GetActiveClusterName()
		request.FailoverVersion = task.GetFailoverVersion()
		request.FailoverNotificationVersion = notificationVersion
	}

	if !recordUpdated {
		return nil
	}

	return h.metadataManagerV2.UpdateNamespace(request)
}

func (h *namespaceReplicationTaskExecutorImpl) validateNamespaceReplicationTask(task *replication.NamespaceTaskAttributes) error {
	if task == nil {
		return ErrEmptyNamespaceReplicationTask
	}

	if task.Id == "" {
		return ErrInvalidNamespaceID
	} else if task.Info == nil {
		return ErrInvalidNamespaceInfo
	} else if task.Config == nil {
		return ErrInvalidNamespaceConfig
	} else if task.ReplicationConfig == nil {
		return ErrInvalidNamespaceReplicationConfig
	}
	return nil
}

func (h *namespaceReplicationTaskExecutorImpl) convertClusterReplicationConfigFromProto(
	input []*commonproto.ClusterReplicationConfiguration) []*persistence.ClusterReplicationConfig {
	output := []*persistence.ClusterReplicationConfig{}
	for _, cluster := range input {
		clusterName := cluster.GetClusterName()
		output = append(output, &persistence.ClusterReplicationConfig{ClusterName: clusterName})
	}
	return output
}

func (h *namespaceReplicationTaskExecutorImpl) convertNamespaceStatusFromProto(input enums.NamespaceStatus) (int, error) {
	switch input {
	case enums.NamespaceStatusRegistered:
		return persistence.NamespaceStatusRegistered, nil
	case enums.NamespaceStatusDeprecated:
		return persistence.NamespaceStatusDeprecated, nil
	default:
		return 0, ErrInvalidNamespaceStatus
	}
}
