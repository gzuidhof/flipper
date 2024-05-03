package resource

import "sort"

// Resources is a list of resources.
type Resources []Resource

// Resource is a common interface for cloud/infra resources.
type Resource interface {
	// ID returns the unique identifier of the resource.
	// Currently as we only support Hetzner, this is the Hetzner ID converted to a string.
	ID() string

	// Equal returns true if the two resources are equal.
	Equal(Resource) bool

	// Name returns the name of the resource.
	Name() string
}

// SortByName sorts the resources by name.
func (r Resources) SortByName() {
	sort.Slice(r, func(i, j int) bool {
		return r[i].Name() < r[j].Name()
	})
}
