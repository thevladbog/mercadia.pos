package app

import "mercadia.dev/pos/services/central-backend/internal/domain"

type CentralPermission string

const (
	PermissionReportingRead        CentralPermission = "reporting.read"
	PermissionReportingCentralRead CentralPermission = "reporting.central.read"
	PermissionUsersManage          CentralPermission = "users.manage"
)

var centralRolePermissions = map[domain.CentralRole][]CentralPermission{
	domain.CentralRoleViewer: {
		PermissionReportingRead,
		PermissionReportingCentralRead,
	},
	domain.CentralRoleAdmin: {
		PermissionReportingRead,
		PermissionReportingCentralRead,
		PermissionUsersManage,
	},
}

func HasCentralPermission(roles []domain.CentralRole, permission CentralPermission) bool {
	for _, role := range roles {
		for _, candidate := range centralRolePermissions[role] {
			if candidate == permission {
				return true
			}
		}
	}
	return false
}

func CheckCentralPermission(roles []domain.CentralRole, permission CentralPermission) error {
	if HasCentralPermission(roles, permission) {
		return nil
	}
	return ErrPermissionDenied
}
