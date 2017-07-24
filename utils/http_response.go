package utils

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/gazoon/bot_libs/logging"
)

const (
	IncorrectRequestData = "INCORRECT_REQUEST_DATA"
	UnknownError         = "UNKNOWN_ERROR"
)

type ErrorData struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
}

type PaginationData struct {
	TotalCount int         `json:"total_count"`
	Data       interface{} `json:"data"`
}

func WriteJSON(ctx context.Context, w http.ResponseWriter, resp interface{}) {
	logger := logging.FromContextAndBase(ctx, gLogger)
	b, err := json.Marshal(resp)
	if err != nil {
		logger.WithField("resp", resp).Errorf("Cannot serialize response to json: %s", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(b)
	if err != nil {
		logger.Errorf("Cannot write json response: %s", err)
	}

}

func ErrorResponse(ctx context.Context, w http.ResponseWriter, errorCode string) {
	ErrorResponseWithMsg(ctx, w, errorCode, errorCode)
}

func InternalErrorResponse(ctx context.Context, w http.ResponseWriter) {
	ErrorResponse(ctx, w, UnknownError)
}

func ErrorResponseWithMsg(ctx context.Context, w http.ResponseWriter, errorCode, msg string) {
	errObj := &ErrorData{Code: errorCode, Msg: msg}
	resp := map[string]interface{}{"error": errObj}
	WriteJSON(ctx, w, resp)
}

func Response(ctx context.Context, w http.ResponseWriter, result interface{}) {
	resp := map[string]interface{}{"result": result}
	WriteJSON(ctx, w, resp)
}

func forceEmptyList(data interface{}) interface{} {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		if v.Len() == 0 {
			return []interface{}{}
		}
	}
	return data
}

func PaginationResponse(ctx context.Context, w http.ResponseWriter, totalCount int, data interface{}) {
	data = forceEmptyList(data)
	Response(ctx, w, &PaginationData{TotalCount: totalCount, Data: data})
}
