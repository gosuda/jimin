// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package database

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type RandflakeNode struct {
	ID          int64  `json:"id"`
	RangeStart  int64  `json:"range_start"`
	RangeEnd    int64  `json:"range_end"`
	ValidFrom   int64  `json:"valid_from"`
	ValidTo     int64  `json:"valid_to"`
	LeaseHolder string `json:"lease_holder"`
}

type User struct {
	ID            int64              `json:"id"`
	Name          string             `json:"name"`
	Email         string             `json:"email"`
	EmailVerified bool               `json:"email_verified"`
	CreatedAt     pgtype.Timestamptz `json:"created_at"`
	UpdatedAt     pgtype.Timestamptz `json:"updated_at"`
}

type WsMember struct {
	ID        int64              `json:"id"`
	WsID      int64              `json:"ws_id"`
	UserID    int64              `json:"user_id"`
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
