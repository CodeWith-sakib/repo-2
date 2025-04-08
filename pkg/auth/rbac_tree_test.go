package auth

import (
	"testing"
)

func TestRoleHierarchyResolver(t *testing.T) {
	resolver := NewRoleHierarchyResolver()

	if !resolver.HasPermission(string(RoleViewer), PermWorkflowRead) {
		t.Error("viewer should have workflow:read")
	}
	if resolver.HasPermission(string(RoleViewer), PermWorkflowWrite) {
		t.Error("viewer should not have workflow:write")
	}

	// Operator inherits viewer permissions
	if !resolver.HasPermission(string(RoleOperator), PermWorkflowRead) {
		t.Error("operator should inherit workflow:read")
	}
	if !resolver.HasPermission(string(RoleOperator), PermRunCreate) {
		t.Error("operator should have run:create")
	}

	// Admin inherits operator and viewer permissions
	if !resolver.HasPermission(string(RoleAdmin), PermWorkflowRead) {
		t.Error("admin should inherit workflow:read")
	}
	if !resolver.HasPermission(string(RoleAdmin), PermRunCreate) {
		t.Error("admin should inherit run:create")
	}
	if !resolver.HasPermission(string(RoleAdmin), PermSystemAdmin) {
		t.Error("admin should have system:admin")
	}
}
