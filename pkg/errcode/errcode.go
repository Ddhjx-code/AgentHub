package errcode

type ErrCode struct {
	Code    int
	Message string
}

func (e *ErrCode) Error() string {
	return e.Message
}

var (
	ErrInvalidParam     = &ErrCode{Code: 400, Message: "invalid parameter"}
	ErrUnauthorized     = &ErrCode{Code: 401, Message: "unauthorized"}
	ErrForbidden        = &ErrCode{Code: 403, Message: "forbidden"}
	ErrNotFound         = &ErrCode{Code: 404, Message: "not found"}
	ErrInternalServer   = &ErrCode{Code: 500, Message: "internal server error"}
	ErrInsufficientFund = &ErrCode{Code: 4001, Message: "insufficient balance"}
	ErrUserBanned       = &ErrCode{Code: 4002, Message: "user is banned"}
	ErrAgentOffline     = &ErrCode{Code: 4003, Message: "agent is offline"}
	ErrEmailExists      = &ErrCode{Code: 4004, Message: "email already registered"}
	ErrInvalidPassword  = &ErrCode{Code: 4005, Message: "incorrect password"}
	ErrTokenExpired     = &ErrCode{Code: 4006, Message: "token expired"}
	ErrTokenInvalid     = &ErrCode{Code: 4007, Message: "invalid token"}
)
