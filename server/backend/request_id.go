package backend

import uuid "github.com/satori/go.uuid"

func NewRequestId() RequestId {
	return RequestId(uuid.Must(uuid.NewV4()).String())
}

type RequestId string

func IsEmptyRequestId(r RequestId) bool {
	return len(r) == 0
}
