/*
Copyright 2022.

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

package controllers

import (
	"context"

	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	log "github.com/sirupsen/logrus"

	"github.com/padok-team/burrito/internal/controllers/terraformlayer"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest"
	"github.com/padok-team/burrito/internal/controllers/terraformrepository"
	"github.com/padok-team/burrito/internal/storage/redis"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/burrito/config"
	//+kubebuilder:scaffold:imports
)

var (
	scheme = runtime.NewScheme()
)

const ()

type Controllers struct {
	config *config.Config
}

type Reconciler interface {
	SetupWithManager(mgr ctrl.Manager) error
}

func New(c *config.Config) *Controllers {
	return &Controllers{
		config: c,
	}
}

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(configv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func (c *Controllers) Exec() {
	log := log.WithContext(context.TODO())

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     c.config.Controller.MetricsBindAddress,
		Port:                   c.config.Controller.KubernetesWehbookPort,
		HealthProbeBindAddress: c.config.Controller.HealthProbeBindAddress,
		LeaderElection:         c.config.Controller.LeaderElection.Enabled,
		LeaderElectionID:       c.config.Controller.LeaderElection.ID,
		// NewCache: func(config *rest.Config, cacheOpts cache.Options) (cache.Cache, error) {
		// 	cacheOpts.Namespaces = c.config.Controller.WatchedNamespaces
		// 	return cache.New(config, cacheOpts)
		// }})
	})
	if err != nil {
		log.Fatalf("unable to start manager: %s", err)
	}

	for _, ctrlType := range c.config.Controller.Types {
		switch ctrlType {
		case "layer":
			if err = (&terraformlayer.Reconciler{
				Client:  mgr.GetClient(),
				Scheme:  mgr.GetScheme(),
				Config:  c.config,
				Storage: redis.New(c.config.Redis.URL, c.config.Redis.Password, c.config.Redis.Database),
			}).SetupWithManager(mgr); err != nil {
				log.Fatalf("unable to create layer controller: %s", err)
			}
			log.Infof("layer controller started successfully")
		case "repository":
			if err = (&terraformrepository.Reconciler{
				Client: mgr.GetClient(),
				Scheme: mgr.GetScheme(),
			}).SetupWithManager(mgr); err != nil {
				log.Fatalf("unable to create repository controller: %s", err)
			}
			log.Infof("repository controller started successfully")
		case "pullrequest":
			if err = (&terraformpullrequest.Reconciler{
				Client: mgr.GetClient(),
				Scheme: mgr.GetScheme(),
				Config: c.config,
			}).SetupWithManager(mgr); err != nil {
				log.Fatalf("unable to create pullrequest controller: %s", err)
			}
			log.Infof("pullrequest controller started successfully")
		default:
			log.Infof("unrecognized controller type %s, ignoring", ctrlType)
		}
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Fatalf("unable to set up health check: %s", err)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Fatalf("unable to set up ready check: %s", err)
	}

	log.Infof("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Fatalf("problem running manager: %s", err)
	}
}
