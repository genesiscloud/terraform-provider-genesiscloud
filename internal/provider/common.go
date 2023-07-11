package provider

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/genesiscloud/genesiscloud-go"
)

type ErrorResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	Error        *genesiscloud.ComputeV1Error
}

var ErrResourceInErrorState = errors.New("the resource is in error state")

func generateErrorMessage(verb string, err error) string {
	return fmt.Sprintf("Error during %s: %s", verb, err)
}

func generateClientErrorMessage(verb string, resp ErrorResponse) string {
	if resp.Error != nil {
		return fmt.Sprintf("Error during %s: [%s] %s %s", verb, resp.HTTPResponse.Status, resp.Error.Code, resp.Error.Message)
	} else {
		body := resp.Body
		if len(body) > 200 {
			body = append(body[:200], []byte("... (truncated)")...)
		}
		return fmt.Sprintf("Error during %s: [%s] %s", verb, resp.HTTPResponse.Status, body)
	}
}

func sliceStringify[T ~string](arr []T) []string {
	ret := make([]string, len(arr))
	for i, value := range arr {
		ret[i] = string(value)
	}

	return ret
}

func pointer[T any](v T) *T {
	return &v
}
