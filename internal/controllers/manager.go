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
	"os"

	logClient "k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	logrusr "github.com/bombsimon/logrusr/v4"
	"github.com/padok-team/burrito/internal/controllers/terraformlayer"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest"
	"github.com/padok-team/burrito/internal/controllers/terraformrepository"
	"github.com/padok-team/burrito/internal/controllers/terraformrun"
	datastore "github.com/padok-team/burrito/internal/datastore/client"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/burrito/config"
	//+kubebuilder:scaffold:imports
)

var (
	scheme = runtime.NewScheme()
)

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
	ctrl.SetLogger(logrusr.New(&log.Logger{
		Out:       os.Stderr,
		Formatter: new(logrus.TextFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.DebugLevel,
	}))
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     c.config.Controller.MetricsBindAddress,
		Port:                   c.config.Controller.KubernetesWebhookPort,
		HealthProbeBindAddress: c.config.Controller.HealthProbeBindAddress,
		LeaderElection:         c.config.Controller.LeaderElection.Enabled,
		LeaderElectionID:       c.config.Controller.LeaderElection.ID,
		Cache: cache.Options{
			Namespaces: append(c.config.Controller.Namespaces, c.config.Controller.MainNamespace),
		},
	})
	if err != nil {
		log.Fatalf("unable to start manager: %s", err)
	}
	datastoreClient := datastore.NewDefaultClient()
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := logClient.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	for _, ctrlType := range c.config.Controller.Types {
		switch ctrlType {
		case "layer":
			if err = (&terraformlayer.Reconciler{
				Client:    mgr.GetClient(),
				Scheme:    mgr.GetScheme(),
				Config:    c.config,
				Recorder:  mgr.GetEventRecorderFor("Burrito"),
				Datastore: datastoreClient,
			}).SetupWithManager(mgr); err != nil {
				log.Fatalf("unable to create layer controller: %s", err)
			}
			log.Infof("layer controller started successfully")
		case "repository":
			if err = (&terraformrepository.Reconciler{
				Client:   mgr.GetClient(),
				Scheme:   mgr.GetScheme(),
				Recorder: mgr.GetEventRecorderFor("Burrito"),
			}).SetupWithManager(mgr); err != nil {
				log.Fatalf("unable to create repository controller: %s", err)
			}
			log.Infof("repository controller started successfully")
		case "run":
			if err = (&terraformrun.Reconciler{
				Client:       mgr.GetClient(),
				Scheme:       mgr.GetScheme(),
				Recorder:     mgr.GetEventRecorderFor("Burrito"),
				Config:       c.config,
				Datastore:    datastoreClient,
				K8SLogClient: clientset,
			}).SetupWithManager(mgr); err != nil {
				log.Fatalf("unable to create run controller: %s", err)
			}
			log.Infof("run controller started successfully")
		case "pullrequest":
			if err = (&terraformpullrequest.Reconciler{
				Client:    mgr.GetClient(),
				Scheme:    mgr.GetScheme(),
				Recorder:  mgr.GetEventRecorderFor("Burrito"),
				Config:    c.config,
				Datastore: datastoreClient,
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
