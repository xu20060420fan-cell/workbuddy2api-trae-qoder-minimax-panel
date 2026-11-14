package ideauth

import (
	"net/http"
	"strings"
	"testing"
)

// 回归（2026-10-24 实测修正）：Qoder 的每日签到活动**只对真实设备指纹下发**。
// 旧策略（每账号派生合成 MachineId、不发送 MachineToken）导致活动列表里永远
// 没有签到活动 —— 表现为"扫描不到签到活动"。正确策略：至少**号池第一个号**用
// 本机真实 runtime-info 指纹（machineToken/machineCode/machineId 三件套）；
// 每设备每日限领一个账号是服务端硬规则，由扫描结果如实说明。
//
// ⚠️ 2026-11-14 更正：本测试原来断言「**所有账号**应使用同一真实设备指纹」
// （da.MachineID == db.MachineID）。那条结论已经错了 —— 它正是「两个号设备一样」
// 的历史来源。用户诉求：「设备是真实设备还是邮箱派生必须手动去调，应该自动
// 生成，反正俩肯定应该不一样」。现在默认 device_mode=auto：
// 号池第一个号 = real（真实设备，与客户端运行时一致），其余号 = derived
// （按邮箱确定性派生，Derived=true、Synthetic=false，Token/Code 照真实发送 ——
// 2026-10-24 那条「必须真实发送」的结论仍然成立）。
func TestQoderDeviceRealFingerprint(t *testing.T) {
	// 「真实指纹」来自本机装的 Qoder 客户端（runtime-info）。CI / 无客户端的机器上
	// 取不到，这时断言真实指纹没有意义（会误报失败），直接跳过；
	// 开发机上有客户端时，下面这些断言照常生效。
	if info := QoderDeviceInfo(); info == nil || info.MachineToken == "" || info.MachineID == "" {
		t.Skip("本机没有可用的 Qoder 客户端指纹（runtime-info），跳过真实设备断言")
	}

	// 号池：a@test.com 是第一个号（→ real），b@test.com 是第二个号（→ derived）。
	dir := t.TempDir()
	t.Setenv("IDEAUTH_POOL_DIR", dir)
	pool := &QoderPoolFile{Accounts: []*QoderSnapshot{
		{Email: "a@test.com", Auth: &QoderAuth{Token: "t1", User: QoderUser{ID: "uid-a", Email: "a@test.com"}}},
		{Email: "b@test.com", Auth: &QoderAuth{Token: "t2", User: QoderUser{ID: "uid-b", Email: "b@test.com"}}},
	}}
	if err := SaveQoderPool(pool); err != nil {
		t.Fatal(err)
	}

	a := &QoderAuth{User: QoderUser{ID: "uid-a", Email: "a@test.com"}}
	b := &QoderAuth{User: QoderUser{ID: "uid-b", Email: "b@test.com"}}

	da := qoderDeviceForAccount(a)
	db := qoderDeviceForAccount(b)
	if da == nil || db == nil {
		t.Fatal("设备对象为空")
	}
	if da.Synthetic {
		t.Error("设备必须按真实指纹发送（Synthetic=false）")
	}
	// 请求头必须带物理设备令牌三件套（活动下发的必要条件）。
	h := http.Header{}
	da.Apply(h)
	if h.Get("Cosy-MachineToken") == "" || h.Get("Cosy-MachineCode") == "" {
		t.Errorf("请求头必须包含物理设备令牌: token=%q code=%q",
			h.Get("Cosy-MachineToken"), h.Get("Cosy-MachineCode"))
	}
	if h.Get("Cosy-MachineId") == "" {
		t.Error("Cosy-MachineId 不能为空")
	}
	// 第一个号用真实设备（Derived=false），第二个号用邮箱派生设备（Derived=true）。
	if da.Derived {
		t.Error("号池第一个号应使用真实设备（Derived=false）")
	}
	if !db.Derived {
		t.Error("号池第二个号应使用邮箱派生设备（Derived=true）")
	}
	// 两个号的设备指纹必须不同（自动分配的核心保证）。
	if da.MachineID == db.MachineID {
		t.Errorf("两个号不应共用同一设备指纹: %s == %s", da.MachineID, db.MachineID)
	}
	if !strings.HasPrefix(da.MachineID, "") && len(da.MachineID) == 0 {
		t.Error("machineId 为空")
	}
}
