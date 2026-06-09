package enum

const (
	StatusEnabled  int16 = 1
	StatusDisabled int16 = 2
)

const (
	PermTypeDirectory int16 = 1
	PermTypeMenu      int16 = 2
	PermTypeButton    int16 = 3
)

const (
	DataScopeAll         int16 = 1
	DataScopeDeptAndSub  int16 = 2
	DataScopeDeptOnly    int16 = 3
	DataScopeSelfOnly    int16 = 4
)

const (
	TenantPlatform uint64 = 0
)

const (
	SystemTypePlatform = "platform"
	SystemTypeShop     = "shop"
)

const (
	ReviewStatusPending  int16 = 1
	ReviewStatusApproved int16 = 2
	ReviewStatusRejected int16 = 3
)

const (
	OrderStatusPending    int16 = 1
	OrderStatusInProgress int16 = 2
	OrderStatusCompleted  int16 = 3
	OrderStatusCancelled  int16 = 4
)
