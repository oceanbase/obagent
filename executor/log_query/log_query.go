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

package log_query

import (
	"regexp"
	"sync/atomic"

	"github.com/oceanbase/obagent/config/mgragent"
	"github.com/oceanbase/obagent/errors"
)

// LogQuery single log query
type LogQuery struct {
	conf                  mgragent.LogQueryConfig
	keywordRegexps        []*regexp.Regexp
	keywords              []string
	excludeKeywordRegexps []*regexp.Regexp
	excludeKeywords       []string
	queryLogParams        *QueryLogRequest
	logEntryChan          chan LogEntry
	count                 int64
}

func NewLogQuery(conf mgragent.LogQueryConfig, queryLogParams *QueryLogRequest, logEntryChan chan LogEntry) (*LogQuery, error) {
	if !queryLogParams.validate() {
		err := errors.New("invalid parameters")
		return nil, err
	}

	var (
		keywordRegexps        []*regexp.Regexp
		excludeKeywordRegexps []*regexp.Regexp
		err                   error
	)
	if queryLogParams.KeywordType == regex {
		keywordRegexps, err = genRegexps(queryLogParams.Keyword)
		if err != nil {
			return nil, err
		}
	}

	if queryLogParams.ExcludeKeywordType == regex {
		excludeKeywordRegexps, err = genRegexps(queryLogParams.ExcludeKeyword)
		if err != nil {
			return nil, err
		}
	}

	return &LogQuery{
		conf:                  conf,
		keywords:              queryLogParams.Keyword,
		keywordRegexps:        keywordRegexps,
		excludeKeywords:       queryLogParams.ExcludeKeyword,
		excludeKeywordRegexps: excludeKeywordRegexps,
		queryLogParams:        queryLogParams,
		logEntryChan:          logEntryChan,
		count:                 0,
	}, nil
}

func genRegexps(regexpStrs []string) ([]*regexp.Regexp, error) {
	keywordRegexps := make([]*regexp.Regexp, 0)
	for _, keyword := range regexpStrs {
		if len(keyword) == 0 {
			continue
		}
		keywordRegexp, err := regexp.Compile(keyword)
		if err != nil {
			return nil, err
		}
		keywordRegexps = append(keywordRegexps, keywordRegexp)
	}
	return keywordRegexps, nil
}

func (l *LogQuery) IncCount() {
	atomic.AddInt64(&l.count, 1)
}

func (l *LogQuery) GetCount() int64 {
	return atomic.LoadInt64(&l.count)
}

func (l *LogQuery) GetLimit() int64 {
	if l.queryLogParams == nil {
		return 0
	}
	return l.queryLogParams.Limit
}

func (l *LogQuery) SendLogEntry(logEntry LogEntry) {
	l.logEntryChan <- logEntry
	l.IncCount()
}

func (l *LogQuery) IsExceedLimit() bool {
	limit := l.GetLimit()
	return limit != 0 && l.GetCount() >= limit
}
