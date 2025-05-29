package errorx

import "net/http"

// NewBadRequest new BadRequest error
// that is mapped to a 400 response.
func NewBadRequest() *Error {
	return New(http.StatusBadRequest, "请求参数错误")
}

// NewUnauthorized new Unauthorized error
// that is mapped to a 401 response.
func NewUnauthorized() *Error {
	return New(http.StatusUnauthorized, "未授权")
}

// NewForbidden new Forbidden error
// that is mapped to a 403 response.
func NewForbidden() *Error {
	return New(http.StatusForbidden, "禁止访问")
}

// ErrNotFound new NotFound error
// that is mapped to a 404 response.
func NewNotFound() *Error {
	return New(http.StatusNotFound, "没有找到,资源不存在")
}

// NewMethodNotAllowed new method not allowed error
// that is mapped to a 405 response.
func NewMethodNotAllowed() *Error {
	return New(http.StatusMethodNotAllowed, "方法不允许")
}

// NewRequestTimeout new request timeout error
// that is mapped to a 408 response.
func NewRequestTimeout() *Error {
	return New(http.StatusRequestTimeout, "请求超时")
}

// NewConflict new Conflict error
// that is mapped to a 409 response.
func NewConflict() *Error {
	return New(http.StatusConflict, "资源冲突")
}

// NewInternalServer new internal server error
// that is mapped to 500 response.
func NewInternalServer() *Error {
	return New(http.StatusInternalServerError, "服务器错误")
}

// NewNotImplemented new not implemented error
// that is mapped to 501 response.
func NewNotImplemented() *Error {
	return New(http.StatusNotImplemented, "未实现")
}

// NewBadGateway new bad gateway error
// that is mapped to 502 response.
func NewBadGateway() *Error {
	return New(http.StatusBadGateway, "网关错误")
}

// NewServiceUnavailable new ServiceUnavailable error
// that is mapped to a HTTP 503 response.
func NewServiceUnavailable() *Error {
	return New(http.StatusServiceUnavailable, "服务器不可用")
}

// NewGatewayTimeout new GatewayTimeout error
// that is mapped to a HTTP 504 response.
func NewGatewayTimeout() *Error {
	return New(http.StatusGatewayTimeout, "网关超时")
}

// NewClientClosed new ClientClosed error
// that is mapped to a HTTP 499 response.
func NewClientClosed() *Error {
	return New(499, "客户端关闭")
}
