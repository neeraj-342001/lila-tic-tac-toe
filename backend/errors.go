package main

import "github.com/heroiclabs/nakama-common/runtime"

var (
	errInternalError  = runtime.NewError("internal server error", 13)
	errMarshal        = runtime.NewError("cannot marshal response", 13)
	errNoUserID       = runtime.NewError("not authenticated", 16) // UNAUTHENTICATED
	errUnmarshal      = runtime.NewError("invalid request payload", 3)
)
