package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SettingType represents the type of setting value
type SettingType string

const (
	SettingTypeString SettingType = "string"
	SettingTypeNumber SettingType = "number"
	SettingTypeBool   SettingType = "boolean"
	SettingTypeJSON   SettingType = "json"
)

// Setting represents a tenant-specific configuration setting
type Setting struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TenantID    string             `bson:"tenant_id" json:"tenantId"`
	Key         string             `bson:"key" json:"key"`
	Value       interface{}        `bson:"value" json:"value"`
	Type        SettingType        `bson:"type" json:"type"`
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	IsPublic    bool               `bson:"is_public" json:"isPublic"` // If true, visible to non-admins

	// Timestamps
	CreatedAt time.Time `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time `bson:"updated_at" json:"updatedAt"`
	UpdatedBy string    `bson:"updated_by,omitempty" json:"updatedBy,omitempty"`
}

// Default setting keys
const (
	SettingRequireApproval      = "require_approval"
	SettingAllowAnonymous       = "allow_anonymous"
	SettingMaxTitleLength       = "max_title_length"
	SettingMaxDescriptionLength = "max_description_length"
	SettingMaxAttachments       = "max_attachments"
	SettingAllowedFileTypes     = "allowed_file_types"
	SettingMaxCommentLength     = "max_comment_length"
	SettingMaxCommentDepth      = "max_comment_depth"
	SettingEnableVoting         = "enable_voting"
	SettingEnableComments       = "enable_comments"
	SettingDefaultStatus        = "default_status"
	SettingStatusOptions        = "status_options"
	SettingBrandingLogo         = "branding_logo"
	SettingBrandingColor        = "branding_color"
	SettingBrandingName         = "branding_name"
	SettingNotifyOnNewFeedback  = "notify_on_new_feedback"
	SettingNotifyAdminEmails    = "notify_admin_emails"
)

// DefaultSettings returns default settings for a new tenant
func DefaultSettings(tenantID string) []Setting {
	now := time.Now()
	return []Setting{
		{TenantID: tenantID, Key: SettingRequireApproval, Value: false, Type: SettingTypeBool, IsPublic: true, CreatedAt: now, UpdatedAt: now},
		{TenantID: tenantID, Key: SettingAllowAnonymous, Value: false, Type: SettingTypeBool, IsPublic: true, CreatedAt: now, UpdatedAt: now},
		{TenantID: tenantID, Key: SettingMaxTitleLength, Value: 200, Type: SettingTypeNumber, IsPublic: true, CreatedAt: now, UpdatedAt: now},
		{TenantID: tenantID, Key: SettingMaxDescriptionLength, Value: 10000, Type: SettingTypeNumber, IsPublic: true, CreatedAt: now, UpdatedAt: now},
		{TenantID: tenantID, Key: SettingMaxAttachments, Value: 5, Type: SettingTypeNumber, IsPublic: true, CreatedAt: now, UpdatedAt: now},
		{TenantID: tenantID, Key: SettingMaxCommentLength, Value: 2000, Type: SettingTypeNumber, IsPublic: true, CreatedAt: now, UpdatedAt: now},
		{TenantID: tenantID, Key: SettingMaxCommentDepth, Value: 3, Type: SettingTypeNumber, IsPublic: true, CreatedAt: now, UpdatedAt: now},
		{TenantID: tenantID, Key: SettingEnableVoting, Value: true, Type: SettingTypeBool, IsPublic: true, CreatedAt: now, UpdatedAt: now},
		{TenantID: tenantID, Key: SettingEnableComments, Value: true, Type: SettingTypeBool, IsPublic: true, CreatedAt: now, UpdatedAt: now},
		{TenantID: tenantID, Key: SettingDefaultStatus, Value: "pending", Type: SettingTypeString, IsPublic: false, CreatedAt: now, UpdatedAt: now},
		{TenantID: tenantID, Key: SettingNotifyOnNewFeedback, Value: true, Type: SettingTypeBool, IsPublic: false, CreatedAt: now, UpdatedAt: now},
	}
}
