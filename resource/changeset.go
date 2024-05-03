package resource

import "fmt"

// Changeset represents a set of changes between two slices of resources.
type Changeset[ResourceType Resource] struct {
	// Added is a list of resources that were added. The order of the list is not guaranteed.
	Added []ResourceType
	// Removed is a list of resources that were removed. The order of the list is not guaranteed.
	Removed []ResourceType
	// Updated is a list of resources that were updated. The order of the list is not guaranteed.
	Updated []ResourceType
}

// GroupChangeset represents the changes between two groups.
type GroupChangeset struct {
	Servers     Changeset[Server]
	FloatingIPs Changeset[FloatingIP]
}

// NewChangeset creates a new changeset, comparing two slices of resources.
// The changeset will contain the resources that were added, removed, or updated.
func NewChangeset[T Resource](oldValues, newValues []T) Changeset[T] {
	changeset := Changeset[T]{
		Added:   make([]T, 0),
		Removed: make([]T, 0),
		Updated: make([]T, 0),
	}

	oldMap := make(map[string]T, len(oldValues))
	for _, r := range oldValues {
		oldMap[r.ID()] = r
	}

	newMap := make(map[string]T, len(newValues))
	for _, r := range newValues {
		newMap[r.ID()] = r
	}

	for _, r := range oldValues {
		if _, ok := newMap[r.ID()]; !ok {
			changeset.Removed = append(changeset.Removed, r)
		}
	}

	for _, r := range newValues {
		old, ok := oldMap[r.ID()]
		if !ok {
			changeset.Added = append(changeset.Added, r)
		} else if !r.Equal(old) {
			changeset.Updated = append(changeset.Updated, r)
		}
	}

	return changeset
}

// Empty returns true if the changeset has no changes.
func (c Changeset[ResourceType]) Empty() bool {
	return len(c.Added) == 0 && len(c.Removed) == 0 && len(c.Updated) == 0
}

// String returns a string representation of the changeset.
func (c Changeset[ResourceType]) String() string {
	return fmt.Sprintf("Changeset{Added: %v, Removed: %v, Updated: %v}", c.Added, c.Removed, c.Updated)
}

// NewGroupChangeset creates a new group changeset, comparing two groups.
func NewGroupChangeset(oldGroup, newGroup Group) GroupChangeset {
	return GroupChangeset{
		Servers:     NewChangeset(oldGroup.Servers, newGroup.Servers),
		FloatingIPs: NewChangeset(oldGroup.FloatingIPs, newGroup.FloatingIPs),
	}
}

// Empty returns true if the changeset is empty.
func (c *GroupChangeset) Empty() bool {
	return c.Servers.Empty() && c.FloatingIPs.Empty()
}

// String returns a string representation of the group changeset.
func (c *GroupChangeset) String() string {
	return fmt.Sprintf("GroupChangeset{Servers: %s, FloatingIPs: %s}", c.Servers, c.FloatingIPs)
}

// IsUpdatesOnly returns true if the changeset only contains updated resources.
func (c *GroupChangeset) IsUpdatesOnly() bool {
	return len(c.Servers.Added) == 0 && len(c.Servers.Removed) == 0 &&
		len(c.FloatingIPs.Added) == 0 && len(c.FloatingIPs.Removed) == 0
}
