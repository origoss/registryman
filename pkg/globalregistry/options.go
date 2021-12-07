package globalregistry

// ReplicationType is a type needed to specify wether the
// replication gets handled by Skopeo or at registry level.
type ReplicationType string

const (
	RegistryReplication ReplicationType = "registry"
	SkopeoReplication   ReplicationType = "skopeo"
)

// CanForceDelete interface describes an option that is needed to
// be able to delete a project when it has repositories in it.
type CanForceDelete interface {
	ForceDeleteProjects() bool
}

// CanReplicate interface describes an option that is needed to
// be able to replicate a project and its contents.
type CanReplicate interface {
	SupportsProjectReplication() ReplicationType
}

// RegistryOptions interface describes the registry options
// coming from CLI options, or from the registry description.
type RegistryOptions interface {
}
