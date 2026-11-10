package ideauth

import (
	"os"
	"strings"
	"testing"
)

// TestQoderRuntimeInfoEnv 国际版/国内版环境码必须是 3 / 0。
//
// ⚠️ 回归防护（2026-11-10 真实 bug）：面板此前没传环境码、stdin 也是空的，
// 于是 runtime-info 落到**国内版(prod)默认**身份，与客户端运行时的
// 国际版身份不是同一套 —— 上游按设备身份过滤每日额度，身份对不上就
// **静默**不下发（不报错、不 401），表现为「客户端里有签到任务，
// 面板却说没下发」。这个断言锁住环境码映射。
func TestQoderRuntimeInfoEnv(t *testing.T) {
	if got := qoderRuntimeInfoEnv("intl"); got != "3" {
		t.Fatalf("国际版环境码 = %q，期望 3", got)
	}
	if got := qoderRuntimeInfoEnv(""); got != "3" {
		t.Fatalf("默认环境码 = %q，期望 3", got)
	}
	if got := qoderRuntimeInfoEnv("CN"); got != "0" {
		t.Fatalf("国内版环境码 = %q，期望 0", got)
	}
	if got := qoderRuntimeInfoEnv("cn"); got != "0" {
		t.Fatalf("国内版环境码 = %q，期望 0", got)
	}
}

// TestQoderRuntimeInfoIdentity 国际版身份必须与客户端运行时一致。
//
// 只在装了 Qoder 客户端的机器上跑（没装就跳过）。断言两件事：
//  1. 环境码 3 能取到指纹，且 **不等于** env=prod（无环境码）取到的那套；
//  2. env=3 取到的 machineToken / machineCode 与「客户端运行时」一致 ——
//     客户端运行时那一套可用 net-log 抓；这里改用「与 env=0 不同」来保证
//     不会退回国内版身份（这正是原 bug 的形态）。
func TestQoderRuntimeInfoIdentity(t *testing.T) {
	exe := QoderRuntimeInfoExe()
	if exe == "" {
		t.Skip("未安装 Qoder 客户端，跳过")
	}
	if _, err := os.Stat(exe); err != nil {
		t.Skip("runtime-info.exe 不可用，跳过")
	}
	intl := qoderRunRuntimeInfo(exe, "3", "")
	cn := qoderRunRuntimeInfo(exe, "0", "")
	if intl == nil || cn == nil {
		t.Skip("runtime-info 无输出，跳过")
	}
	it := qoderStr(intl, "machineToken")
	ct := qoderStr(cn, "machineToken")
	if it == "" {
		t.Skip("未取到 machineToken，跳过")
	}
	if it == ct {
		t.Fatalf("国际版与国内版身份相同（machineToken=%s）—— 环境码没生效", it)
	}
	ic := qoderStr(intl, "machineCode")
	if ic == "" {
		t.Fatalf("国际版身份缺 machineCode")
	}
}

func qoderStr(m map[string]any, k string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[k].(string); ok {
		return strings.TrimSpace(v)
	}
	return ""
}
