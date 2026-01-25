package auth

import (
	"testing"
)

func TestRolePermissionMatrix(t *testing.T) {
	m := NewRolePermissionMatrix()

	m.GrantPermission("viewer", "workflow:read")
	m.GrantPermission("editor", "workflow:write")
	m.GrantPermission("admin", "workflow:delete")

	// Inheritances: admin -> editor -> viewer
	m.SetRoleInheritance("editor", "viewer")
	m.SetRoleInheritance("admin", "editor")

	// Viewer can read, cannot write
	if !m.HasPermission("viewer", "workflow:read") {
		t.Error("viewer should have workflow:read")
	}
	if m.HasPermission("viewer", "workflow:write") {
		t.Error("viewer should NOT have workflow:write")
	}

	// Editor can read and write
	if !m.HasPermission("editor", "workflow:read") {
		t.Error("editor should inherit workflow:read")
	}
	if !m.HasPermission("editor", "workflow:write") {
		t.Error("editor should have workflow:write")
	}

	// Admin has all three
	if !m.HasPermission("admin", "workflow:read") ||
		!m.HasPermission("admin", "workflow:write") ||
		!m.HasPermission("admin", "workflow:delete") {
		t.Error("admin should inherit all permissions")
	}
}
