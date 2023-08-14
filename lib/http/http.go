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

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"
)

type Http interface {
	DownloadFile(path string, url string) error
}

var globalHttpClient = &http.Client{}
var httpClientLock = sync.Mutex{}

func GetGlobalHttpClient() *http.Client {
	httpClientLock.Lock()
	defer httpClientLock.Unlock()
	return globalHttpClient
}

type HttpImpl struct {
}

func UnsetSocksProxy() {
	_ = SetSocksProxy("")
}

func SetSocksProxy(proxyAddr string) error {
	httpClientLock.Lock()
	defer httpClientLock.Unlock()
	proxyAddr = strings.TrimSpace(proxyAddr)
	if proxyAddr == "" {
		log.Info("unset http socks proxy")
		globalHttpClient.Transport = nil
		return nil
	}
	dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
	if err != nil {
		return err
	}
	log.Infof("set http socks proxy to %s", proxyAddr)
	globalHttpClient.Transport = &http.Transport{
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			log.Infof("http connect to %s %s via socks proxy %s", network, address, proxyAddr)
			return dialer.Dial(network, address)
		},
		DisableKeepAlives: true,
	}
	return nil
}

func (h HttpImpl) DownloadFile(path string, url string) error {
	log.WithField("url", url).Info("download file from url start")
	targetFile, err := os.Create(path)
	if err != nil {
		return errors.Errorf("failed to download file from url, cannot open file %s: %s", path, err)
	}
	defer targetFile.Close()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return errors.Errorf("failed to download file from url, cannot create request: %s", err)
	}
	resp, err := GetGlobalHttpClient().Do(req)
	if err != nil {
		return errors.Errorf("failed to download file from url, cannot send request: %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		buf := bytes.NewBufferString("")
		_, _ = io.Copy(buf, resp.Body)
		result := buf.String()
		return errors.Errorf("failed to download file from url, error response: %s", result)
	}

	reader := &DownloadFileReader{
		Reader:   resp.Body,
		Total:    resp.ContentLength,
		Current:  0,
		Progress: 0,
	}
	_, err = io.Copy(targetFile, reader)
	if err != nil {
		return errors.Errorf("failed to download file from url, cannot save response: %s", err)
	}
	log.WithField("url", url).Info("download file from url done")
	return nil
}

// DownloadFileReader custom a reader so that we can print the download progress
type DownloadFileReader struct {
	io.Reader
	Total    int64 // total bytes to download
	Current  int64 // current downloaded bytes
	Progress int   // current max progress, range [0, 100]
}

func (r *DownloadFileReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	if err != nil {
		return n, err
	}
	r.Current += int64(n)
	progress := float64(r.Current*10000/r.Total) / 100
	// print log if the progress proceeds 1%
	if int(progress) > r.Progress {
		log.WithFields(log.Fields{
			"total":    r.Total,
			"current":  r.Current,
			"progress": fmt.Sprintf("%.2f%%", progress),
		}).Info("current progress")
		r.Progress = int(progress)
	}
	return n, err
}
