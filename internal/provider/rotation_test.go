package provider

import (
	"testing"
	"time"

	"github.com/tmalldedede/agentbox/internal/apperr"
)

func TestProfileRotator_GetActiveProfile(t *testing.T) {
	profiles := []*AuthProfile{
		{ID: "p1", Priority: 0, IsEnabled: true},
		{ID: "p2", Priority: 1, IsEnabled: true},
		{ID: "p3", Priority: 0, IsEnabled: false}, // disabled
	}

	rotator := NewProfileRotator(profiles, nil)

	// Should return highest priority (lowest number) enabled profile
	active := rotator.GetActiveProfile()
	if active == nil {
		t.Fatal("expected active profile, got nil")
	}
	if active.ID != "p1" {
		t.Errorf("expected p1, got %s", active.ID)
	}
}

func TestProfileRotator_Cooldown(t *testing.T) {
	profiles := []*AuthProfile{
		{ID: "p1", Priority: 0, IsEnabled: true},
		{ID: "p2", Priority: 1, IsEnabled: true},
	}

	cfg := &RotatorConfig{
		DefaultCooldown:  100 * time.Millisecond,
		MaxFailCount:     2,
		ExtendedCooldown: 200 * time.Millisecond,
	}
	rotator := NewProfileRotator(profiles, cfg)

	// Mark p1 as failed
	rotator.MarkFailed("p1", apperr.ReasonRateLimit)

	// p1 should be in cooldown, should return p2
	active := rotator.GetActiveProfile()
	if active == nil {
		t.Fatal("expected active profile, got nil")
	}
	if active.ID != "p2" {
		t.Errorf("expected p2 (p1 in cooldown), got %s", active.ID)
	}

	// Wait for cooldown to expire
	time.Sleep(250 * time.Millisecond)

	// p1 should be available again
	active = rotator.GetActiveProfile()
	if active.ID != "p1" {
		t.Errorf("expected p1 after cooldown, got %s", active.ID)
	}
}

func TestProfileRotator_MarkSuccess(t *testing.T) {
	profiles := []*AuthProfile{
		{ID: "p1", Priority: 0, IsEnabled: true, FailCount: 5},
	}

	rotator := NewProfileRotator(profiles, nil)

	// Mark success should reset fail count
	rotator.MarkSuccess("p1")

	allProfiles := rotator.GetAllProfiles()
	if allProfiles[0].FailCount != 0 {
		t.Errorf("expected fail count 0, got %d", allProfiles[0].FailCount)
	}
	if allProfiles[0].SuccessCount != 1 {
		t.Errorf("expected success count 1, got %d", allProfiles[0].SuccessCount)
	}
}

func TestProfileRotator_AllInCooldown(t *testing.T) {
	profiles := []*AuthProfile{
		{ID: "p1", Priority: 0, IsEnabled: true},
		{ID: "p2", Priority: 1, IsEnabled: true},
	}

	cfg := &RotatorConfig{
		DefaultCooldown:  100 * time.Millisecond,
		MaxFailCount:     3,
		ExtendedCooldown: 200 * time.Millisecond,
	}
	rotator := NewProfileRotator(profiles, cfg)

	// Initially not all in cooldown
	if rotator.AllInCooldown() {
		t.Error("expected not all in cooldown initially")
	}

	// Mark both as failed with a non-rate-limit reason (which doesn't double cooldown)
	rotator.MarkFailed("p1", apperr.ReasonUnknown)
	rotator.MarkFailed("p2", apperr.ReasonUnknown)

	// Now all should be in cooldown
	if !rotator.AllInCooldown() {
		t.Error("expected all in cooldown after marking both failed")
	}

	// Wait for cooldown
	time.Sleep(150 * time.Millisecond)

	// Should no longer be all in cooldown
	if rotator.AllInCooldown() {
		t.Error("expected not all in cooldown after waiting")
	}
}

func TestProfileRotator_ActiveCount(t *testing.T) {
	profiles := []*AuthProfile{
		{ID: "p1", Priority: 0, IsEnabled: true},
		{ID: "p2", Priority: 1, IsEnabled: true},
		{ID: "p3", Priority: 2, IsEnabled: false},
	}

	rotator := NewProfileRotator(profiles, nil)

	count := rotator.ActiveCount()
	if count != 2 {
		t.Errorf("expected 2 active profiles, got %d", count)
	}
}
