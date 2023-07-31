// Mask sensitive values in logs and error responses.
// Refer to com.alipay.ocp.common.util.log.LogContentUtils in ocp-common

package mask

import "regexp"

var commandPasswordPattern = regexp.MustCompile(`(?i)password(=|:)[^\s]*`)
var commandPasswordReplaceTo = "password${1}xxx"

var scriptPasswordPattern = regexp.MustCompile(`(python)(.*?) (-p=?)[^\s]*`)
var scriptPasswordReplaceTo = "$1$2 ${3}xxx"

var mysqlPasswordPattern = regexp.MustCompile(`(mysql|obclient)(.*?) -p[^\s]*`)
var mysqlPasswordReplaceTo = "$1$2 -pxxx"

var mysqlDSNPattern = regexp.MustCompile(`(.+?):(.+?)@tcp(.+)`)
var mysqlDSNReplaceTo = "$1:xxx@tcp$3"

var dumpBackupPattern = regexp.MustCompile(`(access_id|access_key)=[\w\d]*`)
var dumpBackupReplaceTo = "$1=xxx"

func maskCommandPassword(text string) string {
	return commandPasswordPattern.ReplaceAllString(text, commandPasswordReplaceTo)
}

func maskScriptPassword(text string) string {
	return scriptPasswordPattern.ReplaceAllString(text, scriptPasswordReplaceTo)
}

func maskMysqlPassword(text string) string {
	return mysqlPasswordPattern.ReplaceAllString(text, mysqlPasswordReplaceTo)
}

func maskMysqlDSN(text string) string {
	return mysqlDSNPattern.ReplaceAllString(text, mysqlDSNReplaceTo)
}

func maskDumpBackup(text string) string {
	return dumpBackupPattern.ReplaceAllString(text, dumpBackupReplaceTo)
}

var maskFunctions = []func(string) string{
	maskCommandPassword,
	maskScriptPassword,
	maskMysqlPassword,
	maskMysqlDSN,
	maskDumpBackup,
}

func Mask(text string) string {
	for _, fn := range maskFunctions {
		text = fn(text)
	}
	return text
}

func MaskSlice(texts []string) []string {
	result := make([]string, 0)
	for _, text := range texts {
		result = append(result, Mask(text))
	}
	return result
}
