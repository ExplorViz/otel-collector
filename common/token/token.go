// Package token provides utilities for working with landscape tokens.
package token

// A landscape token is the identifier of a visualization landscape within ExplorViz.
// Tokens can be created or deleted by the user within the frontend.
// The token's secret value prevents unauthorized users from writing to a landscape.
type LandscapeToken struct {
	ID     string
	Secret string
}
