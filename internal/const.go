package internal

const (
	ProjectManifest = "project.json"  // project manifest file (configuration for the platform)
	CGIIgnore       = ".cgiignore"    // file with tar --exclude-from patterns for upload/download filter
	ManifestFile    = "manifest.json" // lambda configuration
	SSHKeySize      = 3072            // generated SSH size for git client
)
