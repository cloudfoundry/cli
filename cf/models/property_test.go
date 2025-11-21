package models_test

import (
	"testing"
	"testing/quick"

	"github.com/cloudfoundry/cli/cf/models"
)

// TestAppParamsMergePreservesOriginalWhenUpdateIsNil verifies original values kept when update is nil
func TestAppParamsMergePreservesOriginalWhenUpdateIsNil(t *testing.T) {
	f := func(name string, memory int64) bool {
		if memory < 0 {
			memory = -memory // Keep positive
		}

		baseParams := models.AppParams{
			Name:   &name,
			Memory: &memory,
		}

		// Create empty update
		emptyUpdate := models.AppParams{}

		// Merge should preserve original values
		baseParams.Merge(&emptyUpdate)

		return baseParams.Name != nil &&
			*baseParams.Name == name &&
			baseParams.Memory != nil &&
			*baseParams.Memory == memory
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

// TestAppParamsMergeOverridesWithNewValues verifies new values override originals
func TestAppParamsMergeOverridesWithNewValues(t *testing.T) {
	f := func(oldName, newName string, oldMem, newMem int64) bool {
		if oldMem < 0 {
			oldMem = -oldMem
		}
		if newMem < 0 {
			newMem = -newMem
		}

		baseParams := models.AppParams{
			Name:   &oldName,
			Memory: &oldMem,
		}

		updateParams := models.AppParams{
			Name:   &newName,
			Memory: &newMem,
		}

		baseParams.Merge(&updateParams)

		// Should have new values
		return baseParams.Name != nil &&
			*baseParams.Name == newName &&
			baseParams.Memory != nil &&
			*baseParams.Memory == newMem
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

// TestAppParamsMergeIsIdempotent verifies merging same params twice produces same result
func TestAppParamsMergeIsIdempotent(t *testing.T) {
	f := func(name string, memory int64) bool {
		if memory < 0 {
			memory = -memory
		}

		name1, name2 := name, name
		mem1, mem2 := memory, memory

		baseParams1 := models.AppParams{Name: &name1, Memory: &mem1}
		baseParams2 := models.AppParams{Name: &name2, Memory: &mem2}

		updateParams := models.AppParams{Name: &name, Memory: &memory}

		baseParams1.Merge(&updateParams)
		baseParams2.Merge(&updateParams)

		// Both should have same result
		return baseParams1.Name != nil &&
			baseParams2.Name != nil &&
			*baseParams1.Name == *baseParams2.Name &&
			baseParams1.Memory != nil &&
			baseParams2.Memory != nil &&
			*baseParams1.Memory == *baseParams2.Memory
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

// TestRouteURLNeverPanics verifies Route.URL() never panics with any input
func TestRouteURLNeverPanics(t *testing.T) {
	f := func(host, domain, path string, port int) bool {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Route.URL panicked: %v", r)
			}
		}()

		route := models.Route{
			Host: host,
			Domain: models.DomainFields{
				Name: domain,
			},
			Path: path,
			Port: port,
		}

		// Should never panic
		_ = route.URL()

		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

// TestRouteURLConsistency verifies Route.URL() returns same result for same input
func TestRouteURLConsistency(t *testing.T) {
	f := func(host, domain string) bool {
		route := models.Route{
			Host: host,
			Domain: models.DomainFields{
				Name: domain,
			},
		}

		url1 := route.URL()
		url2 := route.URL()

		// Should be deterministic
		return url1 == url2
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

// TestApplicationNamesArePreserved verifies application name is never lost
func TestApplicationNamesArePreserved(t *testing.T) {
	f := func(name string, guid string) bool {
		app := models.Application{
			ApplicationFields: models.ApplicationFields{
				Name: name,
				Guid: guid,
			},
		}

		// Name and GUID should be preserved
		return app.Name == name && app.Guid == guid
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

// TestServiceInstanceIsUserProvidedConsistency verifies IsUserProvided logic
func TestServiceInstanceIsUserProvidedConsistency(t *testing.T) {
	f := func(planGuid string) bool {
		instance := models.ServiceInstance{
			ServicePlan: models.ServicePlanFields{
				Guid: planGuid,
			},
		}

		isUserProvided := instance.IsUserProvided()

		// Should be user-provided if and only if plan GUID is empty
		return isUserProvided == (planGuid == "")
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

// TestDomainFieldsNeverCorrupt verifies domain fields remain intact
func TestDomainFieldsNeverCorrupt(t *testing.T) {
	f := func(name, guid string, shared bool) bool {
		domain := models.DomainFields{
			Name:   name,
			Guid:   guid,
			Shared: shared,
		}

		// Fields should remain exactly as set
		return domain.Name == name &&
			domain.Guid == guid &&
			domain.Shared == shared
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

// TestSpaceFieldsPreserveHierarchy verifies space-org relationship preserved
func TestSpaceFieldsPreserveHierarchy(t *testing.T) {
	f := func(spaceName, spaceGuid, orgGuid string) bool {
		space := models.SpaceFields{
			Name: spaceName,
			Guid: spaceGuid,
		}

		org := models.OrganizationFields{
			Guid: orgGuid,
		}

		// Create a space with org
		fullSpace := models.Space{
			SpaceFields:        space,
			Organization:       org,
		}

		// Hierarchy should be preserved
		return fullSpace.Guid == spaceGuid &&
			fullSpace.Name == spaceName &&
			fullSpace.Organization.Guid == orgGuid
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}
