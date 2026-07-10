package app

import (
	"context"

	"mercadia.dev/pos/services/store-edge/internal/domain"
)

type Permission string

const (
	PermissionReturnsCreate            Permission = "returns.create"
	PermissionReturnsApprove           Permission = "returns.approve"
	PermissionDiscountApply            Permission = "discount.apply"
	PermissionRecountApprove           Permission = "recount.approve"
	PermissionCredentialsManage        Permission = "credentials.manage"
	PermissionStoreSettingsManage      Permission = "store_settings.manage"
	PermissionShiftCloseConfirm        Permission = "shift.close.confirm"
	PermissionChangeFundRequestFulfill Permission = "change_fund_request.fulfill"
)

var rolePermissions = map[domain.Role][]Permission{
	domain.RoleCashier:       {},
	domain.RoleSeniorCashier: {PermissionReturnsCreate, PermissionReturnsApprove, PermissionDiscountApply, PermissionRecountApprove, PermissionCredentialsManage, PermissionStoreSettingsManage, PermissionShiftCloseConfirm, PermissionChangeFundRequestFulfill},
	domain.RoleAdmin:         {PermissionReturnsCreate, PermissionReturnsApprove, PermissionDiscountApply, PermissionRecountApprove, PermissionCredentialsManage, PermissionStoreSettingsManage, PermissionShiftCloseConfirm, PermissionChangeFundRequestFulfill},
}

type ActorRoleLookup interface {
	FindActorRoles(ctx context.Context, actorID string) ([]domain.Role, error)
}

func HasPermission(roles []domain.Role, permission Permission) bool {
	for _, role := range roles {
		for _, candidate := range rolePermissions[role] {
			if candidate == permission {
				return true
			}
		}
	}
	return false
}

func CheckPermission(roles []domain.Role, permission Permission) error {
	if HasPermission(roles, permission) {
		return nil
	}
	return ErrPermissionDenied
}

func CheckActorPermission(lookup ActorRoleLookup, ctx context.Context, actorID string, permission Permission) error {
	if lookup == nil || actorID == "" {
		return ErrPermissionDenied
	}
	roles, err := lookup.FindActorRoles(ctx, actorID)
	if err != nil {
		return err
	}
	return CheckPermission(roles, permission)
}
