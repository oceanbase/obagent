package system

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateKernelParamName(t *testing.T) {
	assert.Equal(t, true, validateKernelParamName("net.ipv4.tcp_rmem"))
	assert.Equal(t, false, validateKernelParamName("net.ipv4.tcp_rmem && reboot"))
}

func TestParseKernelParam(t *testing.T) {
	kernelParam := parseKernelParam("net.ipv4.tcp_rmem", "4096 87380 16777216")
	assert.Equal(t, "net.ipv4.tcp_rmem", kernelParam.ParamKey)
	assert.Equal(t, []string{"4096", "87380", "16777216"}, kernelParam.ParamValue)

	kernelParam = parseKernelParam("vm.min_free_kbytes", "2097152")
	assert.Equal(t, "vm.min_free_kbytes", kernelParam.ParamKey)
	assert.Equal(t, []string{"2097152"}, kernelParam.ParamValue)

}

func TestLoadMemInfoFromReader(t *testing.T) {
	buf := bytes.NewBufferString(`MemTotal:       50331648 kB
MemFree:        39023508 kB
MemAvailable:   39023508 kB
Buffers:               0 kB
Cached:          9344484 kB
SwapCached:            0 kB
Active:          6579240 kB
Inactive:        4728204 kB
Active(anon):    1963060 kB
Inactive(anon):       24 kB
Active(file):    4616180 kB
Inactive(file):  4728180 kB
Unevictable:           0 kB
Mlocked:               0 kB
SwapTotal:             0 kB
SwapFree:              0 kB
Dirty:               872 kB
Writeback:             0 kB
AnonPages:      137964852 kB
Mapped:          2273020 kB
Shmem:                 0 kB
Slab:                  0 kB
SReclaimable:          0 kB
SUnreclaim:            0 kB
KernelStack:      159184 kB
PageTables:       414324 kB
NFS_Unstable:          0 kB
Bounce:                0 kB
WritebackTmp:          0 kB
CommitLimit:    395752136 kB
Committed_AS:   210490696 kB
VmallocTotal:   34359738367 kB
VmallocUsed:           0 kB
VmallocChunk:          0 kB
HardwareCorrupted:     0 kB
AnonHugePages:         0 kB
ShmemHugePages:        0 kB
ShmemPmdMapped:        0 kB
CmaTotal:              0 kB
CmaFree:               0 kB
HugePages_Total:       0
HugePages_Free:        0
HugePages_Rsvd:        0
HugePages_Surp:        0
Hugepagesize:       2048 kB
DirectMap4k:     1870860 kB
DirectMap2M:    102365184 kB
DirectMap1G:    700448768 kB`)

	memInfo, err := loadMemInfoFromReader(buf)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 48, len(memInfo.InfoPairs))

	for _, infoPair := range memInfo.InfoPairs {
		if infoPair.Key == "Mapped" {
			assert.Equal(t, uint64(2273020), infoPair.Value)
		}
	}
}

func TestValidateCheckLibExistsParam(t *testing.T) {
	assert.Equal(t, true, validateCheckLibExistsParam("libaio"))
	assert.Equal(t, false, validateCheckLibExistsParam("libaio && reboot"))
}

func TestParseLibExistsInfo(t *testing.T) {
	lineCount, err := parseLibExistsInfo("2\n")
	assert.Nil(t, err)
	assert.Equal(t, int64(2), lineCount)
}
