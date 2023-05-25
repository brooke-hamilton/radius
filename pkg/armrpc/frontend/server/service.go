/*
Copyright 2023 The Radius Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package server

import (
	"context"
	"fmt"
	"net/http"

	manager "github.com/project-radius/radius/pkg/armrpc/asyncoperation/statusmanager"
	"github.com/project-radius/radius/pkg/armrpc/authentication"
	"github.com/project-radius/radius/pkg/armrpc/hostoptions"
	"github.com/project-radius/radius/pkg/kubeutil"
	"github.com/project-radius/radius/pkg/ucp/dataprovider"
	qprovider "github.com/project-radius/radius/pkg/ucp/queue/provider"
	"github.com/project-radius/radius/pkg/ucp/ucplog"
	controller_runtime "sigs.k8s.io/controller-runtime/pkg/client"
)

// Service is the base worker service implementation to initialize and start web service.
type Service struct {
	// ProviderName is the name of provider namespace.
	ProviderName string
	// Options is the server hosting options.
	Options hostoptions.HostOptions
	// StorageProvider is the provider of storage client.
	StorageProvider dataprovider.DataStorageProvider
	// OperationStatusManager is the manager of the operation status.
	OperationStatusManager manager.StatusManager
	// ARMCertManager is the certificate manager of client cert authentication.
	ARMCertManager *authentication.ArmCertManager
	// KubeClient is the Kubernetes controller runtime client.
	KubeClient controller_runtime.Client
}

// Init initializes web service.
func (s *Service) Init(ctx context.Context) error {
	logger := ucplog.FromContextOrDiscard(ctx)

	s.StorageProvider = dataprovider.NewStorageProvider(s.Options.Config.StorageProvider)
	qp := qprovider.New(s.ProviderName, s.Options.Config.QueueProvider)
	opSC, err := s.StorageProvider.GetStorageClient(ctx, s.ProviderName+"/operationstatuses")
	if err != nil {
		return err
	}
	reqQueueClient, err := qp.GetClient(ctx)
	if err != nil {
		return err
	}
	s.OperationStatusManager = manager.New(opSC, reqQueueClient, s.ProviderName, s.Options.Config.Env.RoleLocation)

	s.KubeClient, err = kubeutil.NewRuntimeClient(s.Options.K8sConfig)
	if err != nil {
		return err
	}

	// Initialize the manager for ARM client cert validation
	if s.Options.Config.Server.EnableArmAuth {
		s.ARMCertManager = authentication.NewArmCertManager(s.Options.Config.Server.ArmMetadataEndpoint, logger)
		if err := s.ARMCertManager.Start(ctx); err != nil {
			return err
		}
	}

	return nil
}

// Start starts HTTP server.
func (s *Service) Start(ctx context.Context, opt Options) error {
	logger := ucplog.FromContextOrDiscard(ctx)
	ctx = hostoptions.WithContext(ctx, s.Options.Config)

	address := fmt.Sprintf("%s:%d", s.Options.Config.Server.Host, s.Options.Config.Server.Port)
	server, err := New(ctx, opt)
	if err != nil {
		return err
	}

	// Handle shutdown based on the context
	go func() {
		<-ctx.Done()
		// We don't care about shutdown errors
		_ = server.Shutdown(ctx)
	}()

	logger.Info(fmt.Sprintf("listening on: '%s'...", address))
	err = server.ListenAndServe()
	if err == http.ErrServerClosed {
		// We expect this, safe to ignore.
		logger.Info("Server stopped...")
		return nil
	} else if err != nil {
		return err
	}

	logger.Info("Server stopped...")
	return nil
}
