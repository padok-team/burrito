package syncwindow

import (
	"path/filepath"
	"time"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

// Behavior of sync windows:
// If there are no sync windows at all, sync is not blocked.
// If there we are in a deny window, sync is blocked.
// If there are any allow windows:
//   - If we are in an allow window, sync is not blocked (unless we are also in a deny window).
//   - If we are not in an allow window, sync is blocked.

func IsSyncBlocked(syncWindows []configv1alpha1.SyncWindow, layerName string) bool {
	// If there are no sync windows at all, sync is not blocked.
	if len(syncWindows) == 0 {
		return false
	}

	now := time.Now()

	var hasAllow bool
	var allowWindowActive bool

	for _, window := range syncWindows {
		if !isLayerInSyncWindow(window, layerName) {
			continue
		}

		// Track if there's at least one "allow" window defined
		if window.Kind == "allow" {
			hasAllow = true
		}

		if isWindowActive(window, now) {
			switch window.Kind {
			case "deny":
				// If we're in any deny window, block immediately.
				return true
			case "allow":
				// Mark that we're currently in an active allow window.
				allowWindowActive = true
			}
		}
	}

	// If we have at least one allow window:
	//    - If currently in an allow window, sync is allowed.
	//    - Otherwise, it is blocked.
	if hasAllow {
		return !allowWindowActive
	}

	// If we have no allow windows at all, then all we have is deny windows.
	// Since we're here, it means we're not in any deny window, so sync is not blocked.
	return false
}

func isWindowActive(window configv1alpha1.SyncWindow, now time.Time) bool {
	schedule, err := cron.ParseStandard(window.Schedule)
	if err != nil {
		log.Errorf("failed to parse schedule %q: %v", window.Schedule, err)
		return false
	}

	dur, err := time.ParseDuration(window.Duration)
	if err != nil {
		log.Errorf("failed to parse duration %q: %v", window.Duration, err)
		return false
	}

	// Check if 'now' is within this window
	// schedule.Next(X) gives the next time the cron
	// fires after X. So we go back 'dur' to find the last start,
	// and see if 'now' is between that start and start + dur.
	start := schedule.Next(now.Add(-dur))
	return now.After(start) && now.Before(start.Add(dur))
}

func isLayerInSyncWindow(syncWindow configv1alpha1.SyncWindow, layerName string) bool {
	for _, pattern := range syncWindow.Layers {
		if pattern == "*" {
			return true
		}
		match, err := filepath.Match(pattern, layerName)
		if err != nil {
			continue
		}
		if match {
			return true
		}
	}
	return false
}
