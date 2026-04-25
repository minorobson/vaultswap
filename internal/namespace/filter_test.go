package namespace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplyFilters_NoOptions_ReturnsAll(t *testing.T) {
	namespaces := []string{"dev", "staging", "prod"}
	result := ApplyFilters(namespaces, FilterOptions{})
	assert.Equal(t, namespaces, result)
}

func TestApplyFilters_Prefix(t *testing.T) {
	namespaces := []string{"dev-us", "dev-eu", "prod-us", "staging"}
	result := ApplyFilters(namespaces, FilterOptions{Prefix: "dev"})
	assert.Equal(t, []string{"dev-us", "dev-eu"}, result)
}

func TestApplyFilters_Substring(t *testing.T) {
	namespaces := []string{"dev-us", "prod-us", "prod-eu", "staging"}
	result := ApplyFilters(namespaces, FilterOptions{Substring: "us"})
	assert.Equal(t, []string{"dev-us", "prod-us"}, result)
}

func TestApplyFilters_Exclude(t *testing.T) {
	namespaces := []string{"dev", "staging", "prod"}
	result := ApplyFilters(namespaces, FilterOptions{Exclude: []string{"staging", "prod"}})
	assert.Equal(t, []string{"dev"}, result)
}

func TestApplyFilters_PrefixAndExclude(t *testing.T) {
	namespaces := []string{"team-alpha", "team-beta", "team-gamma", "infra"}
	result := ApplyFilters(namespaces, FilterOptions{
		Prefix:  "team",
		Exclude: []string{"team-beta"},
	})
	assert.Equal(t, []string{"team-alpha", "team-gamma"}, result)
}

func TestApplyFilters_EmptyInput(t *testing.T) {
	result := ApplyFilters([]string{}, FilterOptions{Substring: "dev"})
	assert.Empty(t, result)
}

func TestApplyFilters_ExcludeWithWhitespace(t *testing.T) {
	namespaces := []string{"dev", "staging", "prod"}
	result := ApplyFilters(namespaces, FilterOptions{Exclude: []string{" staging "}})
	assert.Equal(t, []string{"dev", "prod"}, result)
}
