// Author: Ratheesh G Kumar — Backend Engineer (Golang)
// Destination: Back-End/neptune-pamm-server/internal/model/request.go
// Role: Domain models — core schema request / approval workflow (TFB Requests)
// Description: A pending action awaiting approval or scheduled execution.

package model

import (
	"encoding/json"
	"time"
)

// RequestType enumerates the 7 request types.
type RequestType string

const (
	RequestCreateInvestor    RequestType = "CreateInvestor"
	RequestCreateMm          RequestType = "CreateMm"
	RequestAttachInvestor    RequestType = "AttachInvestor"
	RequestDetachInvestor    RequestType = "DetachInvestor"
	RequestChangeBalance     RequestType = "ChangeBalance"
	RequestChangeCredit      RequestType = "ChangeCredit"
	RequestChangeFeeAccounts RequestType = "ChangeFeeAccounts"
)

// RequestStatus enumerates the request lifecycle.
type RequestStatus string

const (
	RequestNew       RequestStatus = "New"
	RequestPlanned   RequestStatus = "Planned"
	RequestExecuting RequestStatus = "Executing"
	RequestCompleted RequestStatus = "Completed"
	RequestApproved  RequestStatus = "Approved"
	RequestRejected  RequestStatus = "Rejected"
	RequestCanceled  RequestStatus = "Canceled"
	RequestError     RequestStatus = "Error"
)

// RequestSchedule enumerates execution scheduling.
type RequestSchedule string

const (
	ScheduleInstant RequestSchedule = "Instant"
	SchedulePlanned RequestSchedule = "Planned"
	ScheduleOff     RequestSchedule = "Off"
)

// Request is a pending action awaiting approval / scheduled execution.
type Request struct {
	ID           int64            `db:"id"           json:"id"`
	TenantID     int64            `db:"tenant_id"    json:"tenant_id"`
	Type         RequestType      `db:"type"         json:"type"`
	Status       RequestStatus    `db:"status"       json:"status"`
	Schedule     *RequestSchedule `db:"schedule"     json:"schedule,omitempty"`
	AccountID    *int64           `db:"account_id"   json:"account_id,omitempty"` // subject (may be NULL pre-create)
	Payload      json.RawMessage  `db:"payload"      json:"payload"`              // the *RequestDescription body
	Comment      *string          `db:"comment"       json:"comment,omitempty"`
	AdminComment *string          `db:"admin_comment" json:"admin_comment,omitempty"`
	PlannedTime  *time.Time       `db:"planned_time"  json:"planned_time,omitempty"`
	CreatedAt    time.Time        `db:"created_at"    json:"created_at"`
	ExecutedAt   *time.Time       `db:"executed_at"   json:"executed_at,omitempty"`
	Error        *string          `db:"error"         json:"error,omitempty"`
}
