package globalregistry

type CanForceDelete interface {
	ForceDeleteProjects() bool
}
