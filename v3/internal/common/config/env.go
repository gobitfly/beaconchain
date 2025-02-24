package config

/**
 * Contains primitives and helper functions related to the definition of the development environment.
 * Should not contain anything related to the context of the chain
 *
 */

// Describes the execution environment of the service.
type Environment string

const (
	Development Environment = "development" // Set up and managed by the developer
	Staging     Environment = "staging"     // Deployed and managed for internal use only
	Production  Environment = "production"  // Deployed and managed for external use
)
