package http

import "github.com/oceanbase/obagent/lib/errors"

var (
	EncodeRequestFailedErr     = errors.InvalidArgument.NewCode("api/client", "encode_request_failed")
	DecodeResultFailedErr      = errors.Internal.NewCode("api/client", "decode_result_failed")
	NoApiClientErr             = errors.FailedPrecondition.NewCode("api/client", "no_api_client")
	ApiRequestFailedErr        = errors.Internal.NewCode("api/client", "api_request_failed")
	ApiRequestGotFailResultErr = errors.Internal.NewCode("api/client", "api_request_got_fail_result")
)
