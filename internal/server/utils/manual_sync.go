package utils

import (
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
)

type ManualSyncStatus string

const (
	ManualSyncNone      ManualSyncStatus = "none"
	ManualSyncAnnotated ManualSyncStatus = "annotated"
	ManualSyncPending   ManualSyncStatus = "pending"
)

func GetManualSyncStatus(layer configv1alpha1.TerraformLayer) ManualSyncStatus {
	if layer.Annotations[annotations.SyncNow] == "true" {
		return ManualSyncAnnotated
	}
	// check the IsSyncScheduled condition on layer
	for _, c := range layer.Status.Conditions {
		if c.Type == "IsSyncScheduled" && c.Status == "True" {
			return ManualSyncPending
		}
	}
	return ManualSyncNone
}

func GetManualApplyStatus(layer configv1alpha1.TerraformLayer) ManualSyncStatus {
	if layer.Annotations[annotations.ApplyNow] == "true" {
		return ManualSyncAnnotated
	}
	// check the IsApplyScheduled condition on layer
	for _, c := range layer.Status.Conditions {
		if c.Type == "IsApplyScheduled" && c.Status == "True" {
			return ManualSyncPending
		}
	}
	return ManualSyncNone
}
