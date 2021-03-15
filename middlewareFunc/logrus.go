package middlewareFunc

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
)

func MakeLogEntry(ctx echo.Context) *log.Entry {
	if ctx == nil {
		return log.WithFields(log.Fields{"at": time.Now().Format("2006-01-02 15:04:05")})
	}
	return log.WithFields(log.Fields{
		"at":     time.Now().Format("2006-01-02 15:04:05"),
		"method": ctx.Request().Method,
		"uri":    ctx.Request().URL.String(),
		"ip":     ctx.Request().RemoteAddr,
	})
}

func MiddleWareLogging(ctx echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		MakeLogEntry(c).Info("Incoming Request")
		return ctx(c)
	}
}

func ErrHandler(err error, ctx echo.Context) {
	report, ok := err.(*echo.HTTPError)
	if !ok {
		report.Message = fmt.Sprintf("http error %d - %v", report.Code, report.Message)
	} else {
		report = echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	MakeLogEntry(ctx).Error(report.Message)
	ctx.HTML(report.Code, report.Message.(string))
}
