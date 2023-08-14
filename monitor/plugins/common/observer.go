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

package common

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bluele/gcache"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/lib/mask"
)

var observer *Observer
var observerLock sync.Mutex

type DbConnectionConfig struct {
	Url     string `yaml:"url"`
	MaxOpen int    `yaml:"maxOpen"`
	MaxIdle int    `yaml:"maxIdle"`
}

func (conn *DbConnectionConfig) String() string {
	return mask.Mask(conn.Url)
}

func (dbConfig *DbConnectionConfig) Target() string {
	var startIdx int
	var endIdx int
	for idx, c := range dbConfig.Url {
		if string(c) == "(" {
			startIdx = idx
		}
		if string(c) == "?" {
			break
		}
		endIdx = idx
	}
	return dbConfig.Url[startIdx : endIdx+1]
}

func (dbConfig *DbConnectionConfig) GetDb() (*sqlx.DB, error) {
	db, err := sqlx.Connect("mysql", dbConfig.Url)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("connect db, config %s", dbConfig))
	}
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err = db.PingContext(timeoutCtx)
	if err != nil {
		db.Close()
		return nil, errors.Wrap(err, fmt.Sprintf("ping db, config %s", dbConfig))
	}
	db.SetMaxOpenConns(dbConfig.MaxOpen)
	db.SetMaxIdleConns(dbConfig.MaxIdle)
	return db, nil
}

type ObserverMetaInfo struct {
	Id               int64
	Ip               string
	Port             int64
	Version          string
	StartServiceTime int64
	AllTenantIds     []int64
	Cache            gcache.Cache
}

type Observer struct {
	MetaInfo *ObserverMetaInfo
	DbConfig *DbConnectionConfig
	Db       *sqlx.DB
	Ctx      context.Context
	Cancel   context.CancelFunc
}

func (o *Observer) refreshObserverBasicInfo() error {
	obVersion, err := o.getObVersion()
	if err != nil {
		return errors.Wrap(err, "get observer version")
	}
	o.MetaInfo.Version = obVersion

	//for observerId
	var sql string
	compareObversion4Result, err := CompareVersion(o.MetaInfo.Version, obVersion4)
	if err != nil {
		return errors.Wrap(err, "compare observer version 4.0.0.0")
	}
	if compareObversion4Result < 0 {
		sql = selectObserverId
	} else {
		sql = selectObserverIdForObVersion4
	}
	result := o.Db.QueryRowx(sql, o.MetaInfo.Ip, o.MetaInfo.Port)
	var observerId int64
	err = result.Scan(&observerId)
	if err != nil {
		return errors.Wrap(err, "scan obServerId")
	}
	o.MetaInfo.Id = observerId
	return nil
}

func (o *Observer) getObVersion() (string, error) {
	var variableName, variableValue, obVersion string
	result := o.Db.QueryRow(showObVersion)
	err := result.Scan(&variableName, &variableValue)
	if err != nil {
		return variableValue, errors.Wrap(err, "scan observer version")
	} else {
		obVersion, err = ParseVersionComment(variableValue)
		if err != nil {
			return obVersion, errors.Wrap(err, "parse observer version")
		} else {
			return obVersion, nil
		}
	}
}

func (o *Observer) refreshObserverStartTime() error {
	var sql string
	compareObversion4Result, err := CompareVersion(o.MetaInfo.Version, obVersion4)
	if err != nil {
		return errors.Wrap(err, "compare observer version 4.0.0.0")
	}
	if compareObversion4Result < 0 {
		sql = selectObserverStartTime
	} else {
		sql = selectObserverStartTimeForObVersion4
	}
	result := o.Db.QueryRowx(sql, o.MetaInfo.Ip, o.MetaInfo.Port)
	var observerStartTime int64
	err = result.Scan(&observerStartTime)
	if err != nil {
		return errors.Wrap(err, "scan basic info")
	}
	o.MetaInfo.StartServiceTime = observerStartTime
	return nil
}

func (o *Observer) refreshAllTenantIds() error {
	var sql string
	args := []interface{}{o.MetaInfo.Ip, o.MetaInfo.Port}
	compareObversion4Result, err := CompareVersion(o.MetaInfo.Version, obVersion4)
	if err != nil {
		return errors.Wrap(err, "compare observer version 4.0.0.0")
	}
	if compareObversion4Result < 0 {
		sql = selectAllTenants
		args = append(args, o.MetaInfo.Ip, o.MetaInfo.Port)
	} else {
		sql = selectAllTenantsForObVersion4
	}
	allTenantIds := make([]int64, 0, 4)
	results, err := o.Db.Queryx(sql, args...)
	if err != nil {
		return errors.Wrap(err, "query all tenant ids")
	}
	defer results.Close()
	for results.Next() {
		var tenantId int64
		err = results.Scan(&tenantId)
		if err != nil {
			log.WithError(err).Errorf("scan tenant id failed %v", err)
		} else {
			allTenantIds = append(allTenantIds, tenantId)
		}
	}
	o.MetaInfo.AllTenantIds = allTenantIds
	return nil
}

func (o *Observer) RefreshTask() {
	err := o.refreshObserverBasicInfo()
	if err != nil {
		log.WithError(err).Errorf("refresh basic info failed %v", err)
	}
	err = o.refreshAllTenantIds()
	if err != nil {
		log.WithError(err).Errorf("refresh all tenants failed %v", err)
	}
	err = o.refreshObserverStartTime()
	if err != nil {
		log.WithError(err).Errorf("refresh observer startTime failed %v", err)
	}
}

func (o *Observer) startRefreshTask() {
	// refresh every 30 seconds, no need to config in config file
	ticker := time.NewTicker(time.Second * 30)
	go func() {
		for {
			select {
			case <-ticker.C:
				o.RefreshTask()
			case <-o.Ctx.Done():
				ticker.Stop()
				log.Info("stop refresh task")
				return
			}
		}
	}()
}

func (o *Observer) reloadDb(dbConfig *DbConnectionConfig) error {
	db, err := dbConfig.GetDb()
	if err != nil {
		return errors.Wrap(err, "reload db")
	}

	if o.Db != nil {
		o.Db.Close()
	}

	o.Db = db
	o.RefreshTask()
	return nil
}

func (o *Observer) Init() error {
	db, err := o.DbConfig.GetDb()
	if err != nil {
		return err
	}
	o.Db = db
	o.MetaInfo = &ObserverMetaInfo{}

	row := o.Db.QueryRowx(selectHostIp)
	err = row.Scan(&o.MetaInfo.Ip)
	if err != nil {
		return errors.Wrap(err, "get ip")
	}

	row = o.Db.QueryRowx(selectRpcPort)
	err = row.Scan(&o.MetaInfo.Port)
	if err != nil {
		return errors.Wrap(err, "get port")
	}
	o.RefreshTask()
	o.Ctx, o.Cancel = context.WithCancel(context.Background())
	o.startRefreshTask()
	o.registerPrometheusItems()
	return nil
}

func (o *Observer) registerPrometheusItems() {
	dbAddr := fmt.Sprintf("%s:%d", o.MetaInfo.Ip, o.MetaInfo.Port)
	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace:   "monagent",
		Subsystem:   "oceanbase",
		Name:        "db_pool_idle",
		Help:        "idle connections of db pool for table_input",
		ConstLabels: prometheus.Labels{"ob_addr": dbAddr},
	}, func() float64 { return float64(o.Db.Stats().Idle) })
}

func (o *Observer) Close() error {
	if o.Cancel != nil {
		o.Cancel()
	}
	if o.Db != nil {
		return o.Db.Close()
	}
	return nil
}

// GetObserver define an observer instance globally to refresh basic meta info and provide db query and cache service
// once the observer initialized successfully, the same instance will be returned each time when db connection url is the same
// when db connection url is different from the observer, an error will be returned
func GetObserver(connectionConfig *DbConnectionConfig) (*Observer, error) {
	var err error
	observerLock.Lock()
	if observer == nil {
		ob := &Observer{
			DbConfig: connectionConfig,
		}
		err = ob.Init()
		if err == nil {
			observer = ob
		} else {
			_ = ob.Close()
		}
	}
	observerLock.Unlock()

	if err != nil {
		return nil, err
	}

	// check db config url, return only if url is the same
	if connectionConfig.Url != observer.DbConfig.Url {
		observerLock.Lock()
		err := observer.reloadDb(connectionConfig)
		observerLock.Unlock()
		if err != nil {
			return nil, err
		}
	}
	return observer, nil
}
