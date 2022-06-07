package moncmd

import (
	"opensvc.com/opensvc/core/instance"
	"opensvc.com/opensvc/core/path"
)

type (
	SetSmon struct {
		Path    path.T
		Node    string
		Monitor instance.Monitor
	}

	SmonDeleted struct {
		Path path.T
		Node string
	}

	SmonUpdated struct {
		Path   path.T
		Node   string
		Status instance.Monitor
	}
)