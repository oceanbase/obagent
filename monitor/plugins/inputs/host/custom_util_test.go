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

package host

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProc(t *testing.T) {
	procDiskstats := proc(EnvDiskstats, DiskStats)
	require.Equal(t, "/proc/diskstats", procDiskstats)
}

func TestParseEthtool(t *testing.T) {
	output := "Settings for eth0:\n\tSupported ports: [ TP ]\n\tSupported link modes:   10baseT/Half 10baseT/Full\n" +
		"\t                        100baseT/Half 100baseT/Full\n\t                        1000baseT/Full\n" +
		"\tSupported pause frame use: Symmetric\n\tSupports auto-negotiation: Yes\n" +
		"\tAdvertised link modes:  10baseT/Half 10baseT/Full\n\t                        100baseT/Half 100baseT/Full\n" +
		"\t                        1000baseT/Full\n\tAdvertised pause frame use: Symmetric\n\tAdvertised auto-negotiation: Yes\n" +
		"\tSpeed: 1000Mb/s\n\tDuplex: Full\n\tPort: Twisted Pair\n\tPHYAD: 1\n\tTransceiver: internal\n\tAuto-negotiation: on\n" +
		"\tMDI-X: on (auto)\n\tSupports Wake-on: pumbg\n\tWake-on: d\n\tCurrent message level: 0x00000007 (7)\n" +
		"\t\t\t       drv probe link\n\tLink detected: yes"
	res := parseEthtoolResult(output)
	var aimNumber = float64(1000 * 1000 * 1000)
	require.Equal(t, aimNumber, res)
}

func TestLoadSnmpTable(t *testing.T) {
	var snmpData = "IcmpMsg: InType0 InType3 InType8 InType11 InType13 OutType0 OutType3 OutType8 OutType14\n" +
		"IcmpMsg: 751986 2612934 453091 182 12 453091 123 752145 12\n" +
		"Tcp: RtoAlgorithm RtoMin RtoMax MaxConn ActiveOpens PassiveOpens AttemptFails EstabResets CurrEstab " +
		"InSegs OutSegs RetransSegs InErrs OutRsts InCsumErrors\n" +
		"Tcp: 1 200 120000 -1 740944980 224853208 262035391 38896109 8 62117830903 82743287604 46881713 89349 69945595 0"
	data := []byte(snmpData)
	res := loadSnmpTable(data)
	outSegs, ok := res["TcpOutSegs"]
	require.True(t, ok)
	outSegs, ok = outSegs.(int64)
	require.True(t, ok)
	require.Equal(t, int64(82743287604), outSegs)
}

func TestDiskstatsTable(t *testing.T) {
	var diskstatsData = "8       0 sda 4226454 7183 278900326 2822313 904453463 2088291940 27903739880" +
		" 2631951341 0 56003192 2635080488\n" +
		"   8       1 sda1 41 0 328 18 0 0 0 0 0 18 18\n" +
		"   8       2 sda2 555 10 49674 305 1073 357 32152 1323 0 1272 1627\n" +
		"   8       3 sda3 2228488 2191 84663978 975542 657453612 153140813 8759291056 2087134793 0 30294774 2088043117\n" +
		"   8       4 sda4 45 0 4168 40 0 0 0 0 0 23 40"
	data := []byte(diskstatsData)
	res := loadDiskstatsTable(data)
	sda, ok := res["sda"]
	require.True(t, ok)
	ioTicks, ok := sda["ioTicks"]
	require.True(t, ok)
	ioTicks, ok = ioTicks.(int64)
	require.True(t, ok)
	require.Equal(t, int64(56003192), ioTicks)
}

func TestDiskstatsFilter(t *testing.T) {
	var diskstatsData = "8       0 sda 4226454 7183 278900326 2822313 904453463 2088291940 27903739880" +
		" 2631951341 0 56003192 2635080488\n" +
		"   8       1 sda1 41 0 328 18 0 0 0 0 0 18 18\n" +
		"   8       2 sda2 555 10 49674 305 1073 357 32152 1323 0 1272 1627\n" +
		"   8       3 sda3 2228488 2191 84663978 975542 657453612 153140813 8759291056 2087134793 0 30294774 2088043117\n" +
		"   8       4 sda4 45 0 4168 40 0 0 0 0 0 23 40"
	data := []byte(diskstatsData)
	res := loadDiskstatsTable(data)
	require.Equal(t, 5, len(res))
	_, ok := res["sda1"]
	require.True(t, ok)
}
