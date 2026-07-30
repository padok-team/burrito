package validatingwebhook

import (
	"context"
	"fmt"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	log "github.com/sirupsen/logrus"

	admissionv1 "k8s.io/api/admission/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var ValidRemediationStrategies = map[string]bool{
	string(configv1alpha1.AutoApply):                 true,
	string(configv1alpha1.AutoApplyWithApproval):     true,
	string(configv1alpha1.PlanOnly):                  true,
	string(configv1alpha1.PlanAndApplyWithApproval):  true,
}

type TerraformLayerValidator struct {
	Client  client.Client
	decoder admission.Decoder
}

func (v *TerraformLayerValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	layer := &configv1alpha1.TerraformLayer{}
	err := v.decoder.Decode(req, layer)
	if err != nil {
		log.Errorf("validating-webhook: could not decode request: %s", err)
		return admission.Errored(400, err)
	}

	log.Infof("validating TerraformLayer %s/%s", layer.Namespace, layer.Name)

	// Validate remediationStrategy enum
	strategy := string(layer.Spec.RemediationStrategy)
	if strategy != "" && !ValidRemediationStrategies[strategy] {
		err := fmt.Errorf("invalid remediationStrategy: '%s'. Valid values: autoApply, autoApplyWithApproval, planOnly, planAndApplyWithApproval", strategy)
		log.Errorf("validating-webhook rejected: %s", err)
		return admission.Denied(err.Error())
	}

	// Validate that path and branch are set when repository is configured
	if layer.Spec.Repository.Name != "" {
		if layer.Spec.Path == "" {
			return admission.Denied("path is required when repository is configured")
		}
		if layer.Spec.Branch == "" {
			return admission.Denied("branch is required when repository is configured")
		}
	}

	log.Infof("validating TerraformLayer %s/%s passed", layer.Namespace, layer.Name)
	return admission.Allowed("")
}

func (v *TerraformLayerValidator) InjectDecoder(d admission.Decoder) error {
	v.decoder = d
	return nil
}
