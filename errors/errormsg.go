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
}

func loadBundleMessage(assetName string) {
	asset, _ := bindata.Asset(assetName)
	bundle.MustParseMessageFileBytes(asset, assetName)
}

// GetMessage Get localized error message
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
