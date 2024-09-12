package automationModels

var (
	IntervalPolicy = "IntervalPolicy"
	//DailyPolicy    = "DailyPolicy"
	//WeeklyPolicy   = "WeeklyPolicy"
	//MonthlyPolicy  = "MonthlyPolicy"
	//CronExpression = "CronExpression"
)

// BackupPolicyRequest is the request structure for backup policy

type BackupPolicyRequest struct {
	Get    GetBackupPolicy    `copier:"must,nopanic"`
	List   ListBackupPolicy   `copier:"must,nopanic"`
	Create CreateBackupPolicy `copier:"must,nopanic"`
	Delete DeleteBackupPolicy `copier:"must,nopanic"`
}

type BackupPolicyResponse struct {
	Get    V1BackupPolicy
	List   ListBackupPolicyResponse
	Create V1BackupPolicy
}

type ListBackupPolicyResponse struct {
	BackupPolicies []V1BackupPolicy `copier:"must,nopanic"`
	Pagination     *V1PageBasedPaginationResponse
}

type V1BackupPolicy struct {
	Meta   *V1Meta
	Config *V1BackupPolicyConfig
}

type DeleteBackupPolicy struct {
	Id string `copier:"must,nopanic"`
}

type CreateBackupPolicy struct {
	TenantId string
	Meta     *V1Meta
	Config   *V1BackupPolicyConfig
}

type ListBackupPolicy struct {
	PaginationPageNumber string `copier:"must,nopanic"`
	PaginationPageSize   string `copier:"must,nopanic"`
	SortSortBy           string `copier:"must,nopanic"`
	SortSortOrder        string `copier:"must,nopanic"`
}

type GetBackupPolicy struct {
	Id string `copier:"must,nopanic"`
}

type V1BackupPolicyConfig struct {
	Schedule []V1Schedule
}

type V1Schedule struct {
	IntervalPolicy   *V1IntervalPolicy `copier:"must,nopanic"`
	DailyPolicy      *V1DailyPolicy    `copier:"must,nopanic"`
	WeeklyPolicy     *V1WeeklyPolicy   `copier:"must,nopanic"`
	MonthlyPolicy    *V1MonthlyPolicy  `copier:"must,nopanic"`
	CronExpression   *string           `copier:"must,nopanic"`
	IncrementalCount *string           `copier:"must,nopanic"`
	Retain           *string           `copier:"must,nopanic"`
}

type V1IntervalPolicy struct {
	Minutes *string `copier:"must,nopanic"`
}

type V1DailyPolicy struct {
	Time *string `copier:"must,nopanic"`
}

type V1WeeklyPolicy struct {
	Day  *WeeklyPolicyWeekday `copier:"must,nopanic"`
	Time *string              `copier:"must,nopanic"`
}

type V1MonthlyPolicy struct {
	Date *string `copier:"must,nopanic"`
	Time *string `copier:"must,nopanic"`
}

type WeeklyPolicyWeekday string
