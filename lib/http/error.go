/*
 * Copyright (c) 2023 OceanBase
 * OBAgent is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

package http

import "github.com/oceanbase/obagent/lib/errors"

var (
	EncodeRequestFailedErr     = errors.InvalidArgument.NewCode("api/client", "encode_request_failed")
	DecodeResultFailedErr      = errors.Internal.NewCode("api/client", "decode_result_failed")
	NoApiClientErr             = errors.FailedPrecondition.NewCode("api/client", "no_api_client")
	ApiRequestFailedErr        = errors.Internal.NewCode("api/client", "api_request_failed")
	ApiRequestGotFailResultErr = errors.Internal.NewCode("api/client", "api_request_got_fail_result")
)
