// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package database

import (
	"database/sql/driver"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

type RelationType string

const (
	RelationTypeINCLUDE RelationType = "INCLUDE"
	RelationTypeREWRITE RelationType = "REWRITE"
	RelationTypeOTHER   RelationType = "OTHER"
)

func (e *RelationType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = RelationType(s)
	case string:
		*e = RelationType(s)
	default:
		return fmt.Errorf("unsupported scan type for RelationType: %T", src)
	}
	return nil
}

type NullRelationType struct {
	RelationType RelationType `json:"relation_type"`
	Valid        bool         `json:"valid"` // Valid is true if RelationType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullRelationType) Scan(value interface{}) error {
	if value == nil {
		ns.RelationType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.RelationType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullRelationType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.RelationType), nil
}

type RandflakeNode struct {
	ID          int64       `json:"id"`
	RangeStart  int64       `json:"range_start"`
	RangeEnd    int64       `json:"range_end"`
	LeaseHolder pgtype.UUID `json:"lease_holder"`
	LeaseStart  int64       `json:"lease_start"`
	LeaseEnd    int64       `json:"lease_end"`
}

type User struct {
	ID            int64              `json:"id"`
	Name          string             `json:"name"`
	Email         string             `json:"email"`
	EmailVerified bool               `json:"email_verified"`
	CreatedAt     pgtype.Timestamptz `json:"created_at"`
	UpdatedAt     pgtype.Timestamptz `json:"updated_at"`
}

type UsersAuth struct {
	ID              int64              `json:"id"`
	UserID          int64              `json:"user_id"`
	ProviderID      int64              `json:"provider_id"`
	ProviderSubject string             `json:"provider_subject"`
	AssociatedData  string             `json:"associated_data"`
	CreatedAt       pgtype.Timestamptz `json:"created_at"`
	UpdatedAt       pgtype.Timestamptz `json:"updated_at"`
}

type WsMember struct {
	ID        int64              `json:"id"`
	WsID      int64              `json:"ws_id"`
	UserID    int64              `json:"user_id"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
}

type WsRelation struct {
	ID        int64              `json:"id"`
	WsID      int64              `json:"ws_id"`
	ObjectID  int64              `json:"object_id"`
	Relation  RelationType       `json:"relation"`
	TargetID  int64              `json:"target_id"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
}

type WsRole struct {
	ID        int64              `json:"id"`
	WsID      int64              `json:"ws_id"`
	Name      string             `json:"name"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
}

type WsRoleMember struct {
	ID         int64              `json:"id"`
	WsID       int64              `json:"ws_id"`
	WsRoleID   int64              `json:"ws_role_id"`
	WsMemberID int64              `json:"ws_member_id"`
	CreatedAt  pgtype.Timestamptz `json:"created_at"`
	UpdatedAt  pgtype.Timestamptz `json:"updated_at"`
}

type Wss struct {
	ID        int64              `json:"id"`
	Name      string             `json:"name"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
}
