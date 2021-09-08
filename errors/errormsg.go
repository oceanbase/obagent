// Copyright (c) 2021 OceanBase
// obagent is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//
// http://license.coscl.org.cn/MulanPSL2
//
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package errors

import (
	"encoding/json"
	"fmt"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"

	"github.com/oceanbase/obagent/bindata"
)

const errorsI18nEnResourceFile = "assets/i18n/error/en.json"

var defaultLanguage = language.English
var bundle *i18n.Bundle

func init() {
	bundle = i18n.NewBundle(defaultLanguage)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	loadBundleMessage(errorsI18nEnResourceFile)
	// TODO add more message file here if more language support needed
}

func loadBundleMessage(assetName string) {
	asset := bindata.MustAsset(assetName)
	bundle.MustParseMessageFileBytes(asset, assetName)
}

//GetMessage Get localized error message
func GetMessage(lang language.Tag, errorCode ErrorCode, args []interface{}) string {
	localizer := i18n.NewLocalizer(bundle, lang.String())
	message, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID: errorCode.key,
	})
	if err != nil {
		return errorCode.key
	}
	return fmt.Sprintf(message, args...)
}
