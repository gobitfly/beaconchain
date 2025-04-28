package config

/**
 * Contains primitives and helper functions related to the definition of the development environment.
 * Should not contain anything related to the context of the chain
 *
 */

// Describes the execution environment of the service.
type Environment string

const (
	Development Environment = "local"      // Set up and managed by the developer
	Staging     Environment = "staging"    // Deployed and managed for internal use only
	Production  Environment = "production" // Deployed and managed for external use
	Hybrid      Environment = "hybrid"     // Service runs locally, cloud resources run in the cloud
)

func EnvironmentFromString(envName string) Environment {
	switch envName {
	case string(Development):
		return Development
	case string(Staging):
		return Staging
	case string(Production):
		return Production
	case string(Hybrid):
		return Hybrid
	default:
		return Environment(envName) // fallback for unknown environments
	}
}
