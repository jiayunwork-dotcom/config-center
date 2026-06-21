package models

import (
	"time"
)

type Tenant struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Name            string    `gorm:"size:100;unique;not null" json:"name"`
	DisplayName     string    `gorm:"size:200" json:"display_name"`
	MaxNamespaces   int       `gorm:"not null;default:10" json:"max_namespaces"`
	MaxConfigItems  int       `gorm:"not null;default:1000" json:"max_config_items"`
	MaxVersions     int       `gorm:"not null;default:100" json:"max_versions"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Namespace struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TenantID    uint      `gorm:"not null;index" json:"tenant_id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Group struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TenantID    uint      `gorm:"not null;index" json:"tenant_id"`
	NamespaceID uint      `gorm:"not null;index" json:"namespace_id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ConfigItem struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	TenantID       uint      `gorm:"not null;index" json:"tenant_id"`
	NamespaceID    uint      `gorm:"not null;index" json:"namespace_id"`
	GroupID        uint      `gorm:"not null;index" json:"group_id"`
	Key            string    `gorm:"size:255;not null" json:"key"`
	Value          string    `gorm:"type:text;not null" json:"value"`
	Format         string    `gorm:"size:20;not null;default:'json'" json:"format"`
	Environment    string    `gorm:"size:50;not null;default:'dev';index" json:"environment"`
	Level          string    `gorm:"size:20;not null;default:'group';index" json:"level"`
	Schema         *string   `gorm:"type:jsonb" json:"schema"`
	CurrentVersion int       `gorm:"not null;default:1" json:"current_version"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type ConfigVersion struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	TenantID     uint      `gorm:"not null;index" json:"tenant_id"`
	ConfigItemID uint      `gorm:"not null;index" json:"config_item_id"`
	Version      int       `gorm:"not null" json:"version"`
	Value        string    `gorm:"type:text;not null" json:"value"`
	Operator     string    `gorm:"size:100;not null;default:'system'" json:"operator"`
	ChangeType   string    `gorm:"size:50;not null;default:'update'" json:"change_type"`
	Diff         string    `gorm:"type:text" json:"diff"`
	Description  string    `gorm:"type:text" json:"description"`
	CreatedAt    time.Time `json:"created_at"`
}

type GrayRelease struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	TenantID       uint      `gorm:"not null;index" json:"tenant_id"`
	ConfigItemID   uint      `gorm:"not null;index" json:"config_item_id"`
	TargetVersion  int       `gorm:"not null" json:"target_version"`
	Strategy       string    `gorm:"size:20;not null" json:"strategy"`
	IPList         []string  `gorm:"type:text[];serializer:json" json:"ip_list"`
	Percentage     int       `json:"percentage"`
	Status         string    `gorm:"size:20;not null;default:'pending';index" json:"status"`
	PushedCount    int       `gorm:"not null;default:0" json:"pushed_count"`
	TotalCount     int       `gorm:"not null;default:0" json:"total_count"`
	StartedAt      *time.Time `json:"started_at"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type ClientConnection struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	TenantID      uint      `gorm:"not null;index" json:"tenant_id"`
	NamespaceID   uint      `gorm:"not null;index" json:"namespace_id"`
	ClientID      string    `gorm:"size:100;not null" json:"client_id"`
	IPAddress     string    `gorm:"size:50" json:"ip_address"`
	ConnectType   string    `gorm:"size:20;not null" json:"connect_type"`
	LastPullAt    *time.Time `json:"last_pull_at"`
	LastPushAt    *time.Time `json:"last_push_at"`
	PushLatencyMs int       `json:"push_latency_ms"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Metric struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TenantID    uint      `gorm:"not null;index" json:"tenant_id"`
	NamespaceID uint      `gorm:"not null;index" json:"namespace_id"`
	MetricType  string    `gorm:"size:50;not null" json:"metric_type"`
	Value       float64   `gorm:"not null" json:"value"`
	Timestamp   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;index" json:"timestamp"`
}

type User struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Username      string    `gorm:"size:100;unique;not null" json:"username"`
	PasswordHash  string    `gorm:"size:255;not null" json:"-"`
	CreatedAt     time.Time `json:"created_at"`
}

type UserRole struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	NamespaceID *uint     `gorm:"index" json:"namespace_id"`
	Role        string    `gorm:"size:20;not null" json:"role"`
	CreatedAt   time.Time `json:"created_at"`
}

type AuditLog struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	TenantID     uint      `gorm:"not null;default:1" json:"tenant_id"`
	UserID       *uint     `gorm:"index" json:"user_id"`
	Username     string    `gorm:"size:100" json:"username"`
	Action       string    `gorm:"size:50;not null" json:"action"`
	ResourceType string    `gorm:"size:50;not null" json:"resource_type"`
	ResourceID   *uint     `gorm:"index" json:"resource_id"`
	ResourceName string    `gorm:"size:255" json:"resource_name"`
	OldValue     string    `gorm:"type:text" json:"old_value"`
	NewValue     string    `gorm:"type:text" json:"new_value"`
	IPAddress    string    `gorm:"size:50" json:"ip_address"`
	CreatedAt    time.Time `gorm:"index" json:"created_at"`
}

type PendingApproval struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ApplicantID  uint      `gorm:"not null;index" json:"applicant_id"`
	Applicant    string    `gorm:"size:100;not null" json:"applicant"`
	ConfigItemID uint      `gorm:"not null;index" json:"config_item_id"`
	ConfigKey    string    `gorm:"size:255;not null" json:"config_key"`
	NewValue     string    `gorm:"type:text;not null" json:"new_value"`
	OldValue     string    `gorm:"type:text" json:"old_value"`
	Environment  string    `gorm:"size:50;not null;default:'prod'" json:"environment"`
	Description  string    `gorm:"type:text" json:"description"`
	Status       string    `gorm:"size:20;not null;default:'pending';index" json:"status"`
	ReviewerID   *uint     `gorm:"index" json:"reviewer_id"`
	Reviewer     string    `gorm:"size:100" json:"reviewer"`
	ReviewNote   string    `gorm:"type:text" json:"review_note"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

const (
	RoleAdmin  = "admin"
	RoleEditor = "editor"
	RoleViewer = "viewer"
)

const (
	ActionCreate    = "create"
	ActionUpdate    = "update"
	ActionDelete    = "delete"
	ActionRollback  = "rollback"
	ActionStart     = "start"
	ActionFullPush  = "full_push"
	ActionGrantRole = "grant_role"
	ActionRevokeRole = "revoke_role"
)

const (
	ResourceNamespace = "namespace"
	ResourceGroup     = "group"
	ResourceConfig    = "config"
	ResourceGray      = "gray"
	ResourceUserRole  = "user_role"
)
