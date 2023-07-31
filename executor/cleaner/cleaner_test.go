package cleaner

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/oceanbase/obagent/config/mgragent"
)

func TestObCleaner_DeleteFileByRetentionDays(t *testing.T) {
	tmpDir, err := prepareTestDirTree("tmp1")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	filesToBeCreated := []string{"a.b.log", "a.b.log.1", "b.log.2", "c.d.log.wf.1", "log.wf.1", "f.log.wf.2"}

	for _, filesToBeCreated := range filesToBeCreated {
		_, err := os.OpenFile(filesToBeCreated, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			t.Fatal(err)
		}
	}

	now := time.Now()
	err = os.Chtimes("b.log.2", now, now.Add(-12*time.Hour))
	if err != nil {
		t.Fatal()
	}
	err = os.Chtimes("f.log.wf.2", now, now.AddDate(0, 0, -3))
	if err != nil {
		t.Fatal()
	}
	err = os.Chtimes("a.b.log.1", now, now.AddDate(0, 0, -4))
	if err != nil {
		t.Fatal()
	}

	type args struct {
		dirToClean    string
		fileReg       string
		retentionDays uint64
	}
	tests := []struct {
		description string
		args        args
		wantErr     bool
	}{
		{
			description: "删除修改时间在 1 天前并且文件名匹配 regex 的文件",
			args: args{
				dirToClean:    tmpDir,
				fileReg:       "([a-z]+.)?[a-z]+.log.[0-9]+",
				retentionDays: 1,
			},
		}, {
			description: "删除修改时间在 2 天前并且文件名匹配 regex 的文件",
			args: args{
				dirToClean:    tmpDir,
				fileReg:       "([a-z]+.)?[a-z]+.log.wf.[0-9]+",
				retentionDays: 2,
			},
		}, {
			description: "不匹配任何文件的情况",
			args: args{
				dirToClean:    tmpDir,
				fileReg:       "([a-z]+.)?[a-z]+.l.[0-9]+",
				retentionDays: 2,
			},
		}, {
			description: "传入非法目录",
			args: args{
				dirToClean:    "empty",
				fileReg:       "([a-z]+.)?[a-z]+.l.[0-9]+",
				retentionDays: 2,
			},
			wantErr: true,
		},
	}
	obCleaner := NewObCleaner(nil)
	for _, tt := range tests {
		Convey(tt.description, t, func() {
			err := obCleaner.DeleteFileByRetentionDays(context.Background(), tt.args.dirToClean, tt.args.fileReg, tt.args.retentionDays)
			if !tt.wantErr {
				So(err, ShouldBeNil)

				retentionDaysDuration := time.Duration(tt.args.retentionDays) * time.Hour * 24

				matchedFiles, err := FindFilesByRegexAndMTime(context.Background(), tt.args.dirToClean, tt.args.fileReg, retentionDaysDuration)
				So(err, ShouldBeNil)
				So(len(matchedFiles), ShouldBeZeroValue)
			} else {
				So(err, ShouldNotBeNil)
			}
		})
	}
}

func TestObCleaner_DeleteFileByKeepPercentage(t *testing.T) {
	tmpDir, err := prepareTestDirTree("tmp1")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	filesToBeCreated := []string{"a.b.log.1", "a.b.log.2", "b.log.2", "c.d.log.wf.1", "log.wf.1", "f.log.wf.2"}

	for _, filesToBeCreated := range filesToBeCreated {
		_, err := os.OpenFile(filesToBeCreated, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			t.Fatal(err)
		}
	}

	err = os.Truncate("a.b.log.1", 1024)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Truncate("a.b.log.2", 1024)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	err = os.Chtimes("a.b.log.2", now, now.AddDate(0, 0, -2))
	if err != nil {
		t.Fatal(err)
	}
	err = os.Truncate("b.log.2", 2048)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Chtimes("f.log.wf.2", now, now.AddDate(0, 0, -2))
	if err != nil {
		t.Fatal(err)
	}

	err = os.Truncate("c.d.log.wf.1", 2048)
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		dirToClean     string
		fileRegex      string
		keepPercentage uint64
		remainedFiles  []string
	}
	tests := []struct {
		description string
		args        args
		wantErr     bool
	}{
		{
			description: "删除超过保留比例的文件，（删除掉 a.b.log.2，其他匹配的保留）",
			args: args{
				dirToClean:     tmpDir,
				fileRegex:      "([a-z]+.)?[a-z]+.log.[0-9]+",
				keepPercentage: 75,
				remainedFiles:  []string{"a.b.log.1", "b.log.2"},
			},
		},
		{
			description: "保留全部匹配的文件",
			args: args{
				dirToClean:     tmpDir,
				fileRegex:      "([a-z]+.)?[a-z]+.log.wf.[0-9]+",
				keepPercentage: 100,
				remainedFiles:  []string{"c.d.log.wf.1", "f.log.wf.2"},
			},
		},
		{
			description: "删除全部匹配的文件",
			args: args{
				dirToClean:     tmpDir,
				fileRegex:      "([a-z]+.)?[a-z]+.log.wf.[0-9]+",
				keepPercentage: 0.0,
				remainedFiles:  []string{},
			},
		},
		{
			description: "不匹配任何文件",
			args: args{
				dirToClean:     tmpDir,
				fileRegex:      "([a-z]+.)?[a-z]+.l.w.[0-9]+",
				keepPercentage: 50,
				remainedFiles:  []string{},
			},
		}, {
			description: "非法目录",
			args: args{
				dirToClean:     "empty",
				fileRegex:      "([a-z]+.)?[a-z]+.l.w.[0-9]+",
				keepPercentage: 50,
				remainedFiles:  []string{},
			},
			wantErr: true,
		},
	}
	obCleaner := NewObCleaner(nil)
	for _, tt := range tests {
		Convey(tt.description, t, func() {
			err := obCleaner.DeleteFileByKeepPercentage(context.Background(), tt.args.dirToClean, tt.args.fileRegex, tt.args.keepPercentage)
			if !tt.wantErr {
				So(err, ShouldBeNil)

				matchedFiles, err := FindFilesByRegexAndMTime(context.Background(), tt.args.dirToClean, tt.args.fileRegex, 0)
				So(err, ShouldBeNil)
				So(len(matchedFiles), ShouldEqual, len(tt.args.remainedFiles))
				for i, matchedFile := range matchedFiles {
					So(matchedFile.Path, ShouldEqual, filepath.Join(tt.args.dirToClean, tt.args.remainedFiles[i]))
				}
			} else {
				So(err, ShouldNotBeNil)
			}
		})
	}
}

func TestObCleaner_Clean(t *testing.T) {
	tmpDir, err := prepareTestDirTree("tmp1")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	filesToBeCreated := []string{"a.b.log.1", "a.b.log.2", "b.log.3", "c.d.log.wf.1", "log.wf.1", "f.log.wf.2"}

	for _, filesToBeCreated := range filesToBeCreated {
		_, err := os.OpenFile(filesToBeCreated, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			t.Fatal(err)
		}
	}

	err = os.Truncate("a.b.log.1", 1024)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Truncate("a.b.log.2", 1024)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Truncate("b.log.3", 2048)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now()
	err = os.Chtimes("a.b.log.1", now, now.AddDate(0, 0, -4))
	if err != nil {
		t.Fatal(err)
	}
	err = os.Chtimes("a.b.log.2", now, now.AddDate(0, 0, -2))
	if err != nil {
		t.Fatal(err)
	}

	err = os.Truncate("c.d.log.wf.1", 2048)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Truncate("f.log.wf.2", 1000)
	if err != nil {
		t.Fatal(err)
	}

	cleanerConfig := &mgragent.CleanerConfig{
		LogCleaners: []*mgragent.LogCleanerRules{
			{
				LogName:       "ob_log",
				Path:          tmpDir,
				DiskThreshold: 0,
				Rules: []*mgragent.Rule{
					{
						FileRegex:      "([a-z]+.)?[a-z]+.log.[0-9]+",
						RetentionDays:  3,
						KeepPercentage: 70,
					},
					{
						FileRegex:      "([a-z]+.)?[a-z]+.log.wf.[0-9]+",
						RetentionDays:  1,
						KeepPercentage: 80,
					},
				},
			},
		},
	}
	conf := &mgragent.ObCleanerConfig{
		RunInterval: 300,
		Enabled:     true,
		CleanerConf: cleanerConfig,
	}

	Convey("按照配置执行 clean，最终应该只剩 b.log.3, f.log.wf.2, log.wf.1 这几个文件", t, func() {
		obCleaner := NewObCleaner(conf)
		err = obCleaner.Clean(context.Background())
		So(err, ShouldBeNil)

		matchedFiles, err := FindFilesByRegexAndMTime(context.Background(), tmpDir, "", 0)
		So(err, ShouldBeNil)
		So(len(matchedFiles), ShouldEqual, 3)
		remainedFiles := []string{"b.log.3", "f.log.wf.2", "log.wf.1"}
		for i, matchedFile := range matchedFiles {
			So(matchedFile.Path, ShouldEqual, filepath.Join(tmpDir, remainedFiles[i]))
		}
	})
}

func TestObCleaner_Run(t *testing.T) {
	conf := &mgragent.ObCleanerConfig{
		RunInterval: time.Millisecond,
		Enabled:     true,
	}
	Convey("运行次数应该 >= 10", t, func() {
		obCleaner := NewObCleaner(conf)
		go func() {
			obCleaner.Run(context.Background())
		}()
		for {
			if obCleaner.runCount >= 10 {
				So(obCleaner.runCount, ShouldBeGreaterThanOrEqualTo, 10)
				obCleaner.Stop()
				break
			}
		}
	})

	Convey("init ob cleaner, then update config", t, func() {
		conf := &mgragent.ObCleanerConfig{
			RunInterval: time.Millisecond,
			CleanerConf: nil,
			Enabled:     true,
		}
		err := InitOBCleanerConf(conf)
		So(err, ShouldBeNil)

		conf.RunInterval = time.Millisecond * 2
		err = UpdateOBCleanerConf(conf)
		So(err, ShouldBeNil)

		obCleaner.Stop()
	})

	Convey("init-update ob cleaner config order is wrong", t, func() {
		// update, before init
		obCleaner = nil
		conf := &mgragent.ObCleanerConfig{
			RunInterval: time.Millisecond,
			CleanerConf: nil,
			Enabled:     true,
		}
		err := UpdateOBCleanerConf(conf)
		So(err, ShouldBeNil)

		obCleaner = nil
		// init, after init
		err = InitOBCleanerConf(conf)
		So(err, ShouldBeNil)
		err = InitOBCleanerConf(conf)
		So(err, ShouldNotBeNil)

		obCleaner.Stop()
	})
}
