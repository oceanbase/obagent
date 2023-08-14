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

package mgragent

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/api/common"
	"github.com/oceanbase/obagent/executor/log_query"
)

// QueryLogRequest log query request params
type QueryLogRequest struct {
	StartTime           time.Time `json:"startTime"`
	EndTime             time.Time `json:"endTime"`
	LogType             string    `json:"logType"`
	Keyword             []string  `json:"keyword"`
	KeywordType         string    `json:"keywordType"`
	ExcludeKeyword      []string  `json:"excludeKeyword"`
	ExcludeKeywordType  string    `json:"excludeKeywordType"`
	LogLevel            []string  `json:"logLevel"`
	ReqId               string    `json:"reqId"`
	LastQueryFileId     string    `json:"lastQueryFileId"`
	LastQueryFileOffset int64     `json:"lastQueryFileOffset"`
	Limit               int64     `json:"limit"`
}

// DownloadLogRequest log download request params
type DownloadLogRequest struct {
	StartTime          time.Time `json:"startTime"`
	EndTime            time.Time `json:"endTime"`
	LogType            []string  `json:"logType"`
	Keyword            []string  `json:"keyword"`
	KeywordType        string    `json:"keywordType"`
	ExcludeKeyword     []string  `json:"excludeKeyword"`
	ExcludeKeywordType string    `json:"excludeKeywordType"`
	LogLevel           []string  `json:"logLevel"`
	ReqId              string    `json:"reqId"`
}

type LogEntryResponse struct {
	LogAt      time.Time `json:"logAt"`
	LogLine    string    `json:"logLine"`
	LogLevel   string    `json:"logLevel"`
	FileName   string    `json:"fileName"`
	FileId     string    `json:"fileId"`
	FileOffset int64     `json:"fileOffset"`
}

type QueryLogResponse struct {
	LogEntries []LogEntryResponse `json:"logEntries"`
	FileId     string             `json:"fileId"`
	FileOffset int64              `json:"fileOffset"`
}

func queryLogHandler(c *gin.Context) {
	ctx := common.NewContextWithTraceId(c)
	ctxLog := log.WithContext(ctx)
	var param QueryLogRequest
	err := c.BindJSON(&param)
	if err != nil {
		ctxLog.WithError(err).Error("bindJson failed")
		return
	}
	if log_query.GlobalLogQuerier.GetConf().QueryTimeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, log_query.GlobalLogQuerier.GetConf().QueryTimeout)
		defer cancel()
	}

	ctxLog.WithField("param", param)
	ctxLog.Info("invoke log query")

	// Limit the maximum number of queries at a time
	if param.Limit == 0 {
		param.Limit = 200
	}

	logQueryReqParam, err := buildLogQueryReqParams(param)
	if err != nil {
		ctxLog.WithError(err).Error("buildLogQueryReqParams failed")
		return
	}

	logEntryChan := make(chan log_query.LogEntry, 1)
	logQuery, err := log_query.NewLogQuery(log_query.GlobalLogQuerier.GetConf(), logQueryReqParam, logEntryChan)
	if err != nil {
		ctxLog.WithError(err).Error("create NewLogQuery failed")
		common.SendResponse(c, nil, err)
		return
	}

	logEntries := make([]LogEntryResponse, 0)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for logEntry := range logEntryChan {
			logEntries = append(logEntries, buildLogEntryResp(logEntry))
		}
	}()

	lastPos, err := log_query.GlobalLogQuerier.Query(ctx, logQuery)
	if err != nil {
		ctxLog.WithError(err).Error("query failed")
		common.SendResponse(c, nil, err)
		return
	}
	wg.Wait()

	resp := QueryLogResponse{
		LogEntries: logEntries,
	}
	if lastPos != nil {
		resp.FileId = fmt.Sprintf("%d", lastPos.FileId)
		resp.FileOffset = lastPos.FileOffset
	}
	common.SendResponse(c, resp, nil)
}

func downloadLogHandler(c *gin.Context) {
	ctx := common.NewContextWithTraceId(c)
	ctxLog := log.WithContext(ctx)
	var param DownloadLogRequest
	err := c.BindJSON(&param)
	if err != nil {
		ctxLog.WithError(err).Error("bindJson failed")
		return
	}
	if log_query.GlobalLogQuerier.GetConf().DownloadTimeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, log_query.GlobalLogQuerier.GetConf().DownloadTimeout)
		defer cancel()
	}

	ctxLog.WithField("param", param)
	ctxLog.Info("invoke log download")

	w := c.Writer
	zipWriter := zip.NewWriter(w)

	prevLogFileName := ""
	for _, logType := range param.LogType {
		queryLogReq := QueryLogRequest{
			StartTime:      param.StartTime,
			EndTime:        param.EndTime,
			LogType:        logType,
			Keyword:        param.Keyword,
			ExcludeKeyword: param.ExcludeKeyword,
			LogLevel:       param.LogLevel,
			ReqId:          param.ReqId,
		}

		logQueryReqParam, err := buildLogQueryReqParams(queryLogReq)
		if err != nil {
			ctxLog.WithError(err).Error("buildLogQueryReqParams failed")
			return
		}

		logEntryChan := make(chan log_query.LogEntry, 1)
		logQuery, err := log_query.NewLogQuery(log_query.GlobalLogQuerier.GetConf(), logQueryReqParam, logEntryChan)
		if err != nil {
			ctxLog.WithError(err).Error("create NewLogQuery failed")
			common.SendResponse(c, nil, err)
			return
		}

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			var zipFileWriter io.Writer

			for logEntry := range logEntryChan {
				if logEntry.FileName != prevLogFileName {
					zipFileWriter, err = zipWriter.Create(logEntry.FileName)
					if err != nil {
						ctxLog.WithField("logEntry", logEntry).WithError(err).Warn("zipFileWriter.Create failed")
					}
					prevLogFileName = logEntry.FileName
				}
				logLineBytes := append(logEntry.LogLine, '\n')
				_, err1 := zipFileWriter.Write(logLineBytes)
				if err1 != nil {
					ctxLog.WithField("logEntry", logEntry).WithError(err1).Warn("write log entry bytes failed")
					continue
				}
			}
		}()

		_, err = log_query.GlobalLogQuerier.Query(ctx, logQuery)
		if err != nil {
			ctxLog.WithError(err).Error("query failed")
			common.SendResponse(c, nil, err)
			return
		}
		wg.Wait()
	}
	err = zipWriter.Close()
	if err != nil {
		ctxLog.WithError(err).Error("close failed")
	}
	w.Flush()

	common.SendResponse(c, "END-OF-STREAM", nil)
}

func buildLogQueryReqParams(req QueryLogRequest) (*log_query.QueryLogRequest, error) {
	var (
		lastQueryFileId uint64
		err             error
	)
	if req.LastQueryFileId != "" {
		lastQueryFileId, err = strconv.ParseUint(req.LastQueryFileId, 10, 64)
		if err != nil {
			return nil, err
		}
	}
	return &log_query.QueryLogRequest{
		StartTime:           req.StartTime,
		EndTime:             req.EndTime,
		LogType:             req.LogType,
		Keyword:             req.Keyword,
		KeywordType:         log_query.ConditionType(req.KeywordType),
		ExcludeKeyword:      req.ExcludeKeyword,
		ExcludeKeywordType:  log_query.ConditionType(req.ExcludeKeywordType),
		LogLevel:            req.LogLevel,
		ReqId:               req.ReqId,
		LastQueryFileId:     lastQueryFileId,
		LastQueryFileOffset: req.LastQueryFileOffset,
		Limit:               req.Limit,
	}, nil
}

func buildLogEntryResp(logEntry log_query.LogEntry) LogEntryResponse {
	return LogEntryResponse{
		LogAt:      logEntry.LogAt,
		LogLine:    string(logEntry.LogLine),
		LogLevel:   logEntry.LogLevel,
		FileName:   logEntry.FileName,
		FileId:     fmt.Sprintf("%d", logEntry.FileId),
		FileOffset: logEntry.FileOffset,
	}
}
