// Package checker provides controllers that call health checks and manage the state associated with them.
//
// It's built to be compatible with the check package, which provides the actual implementations of the health checks.
//
// The controllers in this package are composable, with roughly the following usual pattern:
//
// * `Periodic` - runs a health check periodically
// * `Stateful` - runs a health check periodically and keeps track of the state, this uses `Periodic` internally.
// * `StatefulMulti` - combines multiple stateful health checks, this uses `Stateful` internally.
// * `Server` - checks the health of a server's public interfaces, this uses `StatefulMulti` internally.
package checker
