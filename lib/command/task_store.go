package command

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// StatusStore Abstract storage for execution status
type StatusStore interface {
	//Create create a new execution store, and save execution status
	//if the token already exists, it will return an ExecutionAlreadyExistsErr
	Create(token ExecutionToken, execution *Execution) error
	//Store save or update execution status of a token.
	Store(token ExecutionToken, execution *Execution) error
	//Load load execution status by the token
	Load(token ExecutionToken) (StoredStatus, error)
	//Delete delete execution data by the token
	Delete(token ExecutionToken) error
	// CreateStatus create a new execution store, and directly save status
	//if the token already exists, it will return an ExecutionAlreadyExistsErr
	CreateStatus(token ExecutionToken, status StoredStatus) error
	//StoreStatus directly save or update status of a token.
	StoreStatus(token ExecutionToken, status StoredStatus) error
}

// FileTaskStore StatusStore implement that save data in local file system.
// It uses json to serialize/deserialize structured data
type FileTaskStore struct {
	dir    string
	expire time.Duration
}

// NewFileTaskStore create a new FileTaskStore with a fs dir to store data in.
func NewFileTaskStore(dir string) *FileTaskStore {
	return &FileTaskStore{
		dir: dir,
	}
}

// FilePrefix all files stored by FileTaskStore will have a name starts with it.
const FilePrefix = "ocp_agent_task_"

// StoredStatus struct used to serialize/deserialize execution status
type StoredStatus struct {
	ResponseType DataType
	Param        interface{}
	Annotation   map[string]interface{}
	Finished     bool
	Ok           bool
	Result       interface{}
	Err          string
	Progress     interface{}
	StartAt      int64
	EndAt        int64
}

// Path returns file path of the token
func (fts *FileTaskStore) Path(token ExecutionToken) string {
	return filepath.Join(fts.dir, FilePrefix+token.String())
}

func (fts *FileTaskStore) Create(token ExecutionToken, execution *Execution) error {
	status := executionToStoredStatus(execution)
	return fts.store(token, status, true)
}

func (fts *FileTaskStore) Store(token ExecutionToken, execution *Execution) error {
	status := executionToStoredStatus(execution)
	return fts.store(token, status, false)
}

func (fts *FileTaskStore) CreateStatus(token ExecutionToken, status StoredStatus) error {
	return fts.store(token, status, true)
}

func (fts *FileTaskStore) StoreStatus(token ExecutionToken, status StoredStatus) error {
	return fts.store(token, status, false)
}

func (fts *FileTaskStore) store(token ExecutionToken, status StoredStatus, create bool) error {
	filePath := fts.Path(token)
	var flag int
	if create {
		flag = os.O_CREATE | os.O_RDWR | os.O_EXCL
	} else {
		flag = os.O_CREATE | os.O_RDWR | os.O_TRUNC
	}
	f, err := os.OpenFile(filePath, flag, 0644)
	if err != nil {
		if os.IsExist(err) {
			return ExecutionAlreadyExistsErr
		}
		return fmt.Errorf("open task store file failed %v", err)
	}
	defer f.Close()

	err = fts.storeToWriter(status, f)
	if err != nil {
		return fmt.Errorf("write task store data failed %v", err)
	}
	err = f.Sync()
	if err != nil {
		return fmt.Errorf("sync task store data failed %v", err)
	}

	return nil
}

func executionToStoredStatus(execution *Execution) StoredStatus {
	execCtx := execution.ExecutionContext()
	input := execCtx.Input()
	status := execCtx.Output().Status()
	return StoredStatus{
		ResponseType: execution.Command().ResponseType(),
		Param:        input.Param(),
		Annotation:   input.Annotation(),
		Finished:     status.Finished,
		Ok:           status.Ok,
		Result:       status.Result,
		Err:          status.Err,
		Progress:     status.Progress,
		StartAt:      execution.startAt.UnixNano(),
		EndAt:        execution.endAt.UnixNano(),
	}
}

func (fts *FileTaskStore) storeToWriter(status StoredStatus, f io.Writer) error {
	//header := executionToStoredStatus(execution)
	isStructured := status.ResponseType == TypeStructured
	result := status.Result
	if !isStructured {
		status.Result = nil
	}
	headerJson, err := json.Marshal(status)
	if err != nil {
		return err
	}
	_, err = f.Write(headerJson)
	if err != nil {
		return err
	}
	_, err = f.Write([]byte{'\n'})
	if !isStructured {
		if err != nil {
			return err
		}
		switch result.(type) {
		case []byte:
			_, err = f.Write(result.([]byte))
		case string:
			_, err = f.Write([]byte(result.(string)))
		}
	}
	return err
}

func (fts *FileTaskStore) Load(token ExecutionToken) (StoredStatus, error) {
	filePath := fts.Path(token)
	f, err := os.Open(filePath)
	if err != nil {
		return StoredStatus{}, err
	}
	defer f.Close()
	//f.Stat() check size
	return fts.loadFromReader(f)
}

func (fts *FileTaskStore) loadFromReader(f io.Reader) (StoredStatus, error) {
	reader := bufio.NewReader(f)

	headerBytes, err := reader.ReadBytes('\n')
	if err != nil {
		return StoredStatus{}, err
	}
	header := StoredStatus{}
	err = json.Unmarshal(headerBytes, &header)
	if err != nil {
		return StoredStatus{}, err
	}

	if header.ResponseType != TypeStructured {
		body, err := ioutil.ReadAll(reader)
		if err != nil {
			return StoredStatus{}, err
		}
		if header.ResponseType == TypeBinary {
			header.Result = body
		} else if header.ResponseType == TypeText {
			header.Result = string(body)
		}
	}
	return header, nil
}

func (fts *FileTaskStore) Delete(token ExecutionToken) error {
	filePath := fts.Path(token)
	return os.Remove(filePath)
}

// Cleanup removes all stored files which mtime before expire duration ago.
func (fts *FileTaskStore) Cleanup(expire time.Duration) {
	d, err := os.Open(fts.dir)
	if err != nil {
		return
	}
	defer d.Close()
	names, err := d.Readdirnames(0)
	if err != nil {
		return
	}
	now := time.Now()
	for _, name := range names {
		if !strings.HasPrefix(name, FilePrefix) {
			continue
		}
		p := filepath.Join(fts.dir, name)

		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		if info.IsDir() {
			continue
		}
		if now.Sub(info.ModTime()) > expire {
			_ = os.Remove(p)
		}
	}
}
