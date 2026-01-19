package provider

import "errors"

var (
	// ErrProviderIDRequired is returned when provider ID is empty
	ErrProviderIDRequired = errors.New("provider ID is required")

	// ErrProviderNameRequired is returned when provider name is empty
	ErrProviderNameRequired = errors.New("provider name is required")

	// ErrProviderAgentRequired is returned when provider agent is empty
	ErrProviderAgentRequired = errors.New("provider agent is required")

	// ErrProviderNotFound is returned when provider is not found
	ErrProviderNotFound = errors.New("provider not found")
)
