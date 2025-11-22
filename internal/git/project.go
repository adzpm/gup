package git

// Project represents a project that can be cloned
type Project struct {
	// Name is the project name
	Name string
	// PathWithNamespace is the full path including namespace
	PathWithNamespace string
	// HTTPURLToRepo is the HTTP URL for cloning the repository
	HTTPURLToRepo string
}
