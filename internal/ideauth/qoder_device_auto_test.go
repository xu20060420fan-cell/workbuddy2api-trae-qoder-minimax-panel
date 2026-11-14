package ideauth

import (
	"os"
	"path/filepath"
	"testing"
)

// TestQoderAutoDeviceAssignment 锁死「设备身份自动分配」的规则：
//
//   - 号池第一个号 → real（与客户端运行时身份一致）；
//   - 其余号 → derived（按邮箱确定性派生，任意两个号必然不同）。
//
// 用户诉求（2026-11-14）：「设备是真实设备还是邮箱派生必须手动去调，
// 应该自动生成，反正俩肯定应该不一样」。
func TestQoderAutoDeviceAssignment(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("IDEAUTH_POOL_DIR", dir)
	t.Setenv("QODER_EGRESS_FILE", filepath.Join(dir, "qoder_egress.json"))

	pool := &QoderPoolFile{Accounts: []*QoderSnapshot{
		{Email: "first@test.com", Auth: &QoderAuth{Token: "t1", User: QoderUser{Email: "first@test.com"}}},
		{Email: "second@test.com", Auth: &QoderAuth{Token: "t2", User: QoderUser{Email: "second@test.com"}}},
		{Email: "third@test.com", Auth: &QoderAuth{Token: "t3", User: QoderUser{Email: "third@test.com"}}},
	}}
	if err := SaveQoderPool(pool); err != nil {
		t.Fatal(err)
	}

	// 首次补全：号池每个号都自动拿到一条出口配置，device_mode=auto。
	n, err := EnsureQoderEgress()
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Fatalf("首次补全应写入 3 条配置，实际 %d", n)
	}
	eg, err := LoadQoderEgress()
	if err != nil {
		t.Fatal(err)
	}
	for _, email := range []string{"first@test.com", "second@test.com", "third@test.com"} {
		if got := eg.Profiles[email].DeviceMode; got != QoderDeviceModeAuto {
			t.Fatalf("%s 的 device_mode 应为 auto，实际 %q", email, got)
		}
		if eg.Profiles[email].ProfileDir == "" {
			t.Fatalf("%s 的 profile_dir 应自动生成", email)
		}
	}

	// 展开后的生效值：第一个 real，其余 derived —— 任意两个号必然不同。
	want := map[string]string{
		"first@test.com":  QoderDeviceModeReal,
		"second@test.com": QoderDeviceModeDerived,
		"third@test.com":  QoderDeviceModeDerived,
	}
	for email, w := range want {
		if got := QoderEffectiveDeviceMode(email, eg.Profiles[email].DeviceMode); got != w {
			t.Errorf("%s 的生效设备模式应为 %q，实际 %q", email, w, got)
		}
	}

	// 幂等：再跑一次不产生改动。
	if n, err := EnsureQoderEgress(); err != nil || n != 0 {
		t.Fatalf("重复补全应无改动，实际 changed=%d err=%v", n, err)
	}
}

// TestQoderEgressLegacyRealMigratedToAuto 锁死存量迁移：
//
// 老版本默认 device_mode=real，号池里每个号都被写成 real（导致多号共用一台设备）。
// 「用户从未真正改过」的条目（没配代理 + profile 目录就是自动生成的那个）应迁移到 auto；
// 配了代理 / 改过目录名的条目保持不动（尊重手动配置）。
func TestQoderEgressLegacyRealMigratedToAuto(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("IDEAUTH_POOL_DIR", dir)
	egPath := filepath.Join(dir, "qoder_egress.json")
	t.Setenv("QODER_EGRESS_FILE", egPath)

	pool := &QoderPoolFile{Accounts: []*QoderSnapshot{
		{Email: "a@test.com", Auth: &QoderAuth{Token: "t1", User: QoderUser{Email: "a@test.com"}}},
		{Email: "b@test.com", Auth: &QoderAuth{Token: "t2", User: QoderUser{Email: "b@test.com"}}},
	}}
	if err := SaveQoderPool(pool); err != nil {
		t.Fatal(err)
	}

	legacy := &QoderEgressPool{Profiles: map[string]QoderEgressProfile{
		// 未改过：real + 无代理 + 自动生成的 profile 目录 → 迁移到 auto。
		"a@test.com": {
			Email: "a@test.com", DeviceMode: QoderDeviceModeReal,
			ProfileDir: filepath.Join("D:\\data", "qoder-web", QoderProfileDirName("a@test.com")),
		},
		// 手动配了代理 → 不动。
		"b@test.com": {
			Email: "b@test.com", DeviceMode: QoderDeviceModeReal,
			ProxyURL: "http://127.0.0.1:1080",
			ProfileDir: filepath.Join("D:\\data", "qoder-web", QoderProfileDirName("b@test.com")),
		},
	}}
	if err := SaveQoderEgress(legacy); err != nil {
		t.Fatal(err)
	}

	if _, err := EnsureQoderEgress(); err != nil {
		t.Fatal(err)
	}
	eg, err := LoadQoderEgress()
	if err != nil {
		t.Fatal(err)
	}
	if got := eg.Profiles["a@test.com"].DeviceMode; got != QoderDeviceModeAuto {
		t.Errorf("未手动改过的 real 条目应迁移为 auto，实际 %q", got)
	}
	if got := eg.Profiles["b@test.com"].DeviceMode; got != QoderDeviceModeReal {
		t.Errorf("配了代理的条目应保留手动值 real，实际 %q", got)
	}
}

// TestQoderResetEgressAuto 锁死「全部恢复自动」：所有号 device_mode=auto，代理保留。
func TestQoderResetEgressAuto(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("IDEAUTH_POOL_DIR", dir)
	egPath := filepath.Join(dir, "qoder_egress.json")
	t.Setenv("QODER_EGRESS_FILE", egPath)

	pool := &QoderPoolFile{Accounts: []*QoderSnapshot{
		{Email: "a@test.com", Auth: &QoderAuth{Token: "t1", User: QoderUser{Email: "a@test.com"}}},
	}}
	if err := SaveQoderPool(pool); err != nil {
		t.Fatal(err)
	}
	if _, err := SetQoderEgress("a@test.com", QoderDeviceModeDerived, "http://127.0.0.1:1080", ""); err != nil {
		t.Fatal(err)
	}
	n, err := ResetQoderEgressAuto()
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("应改动 1 个账号，实际 %d", n)
	}
	eg, err := LoadQoderEgress()
	if err != nil {
		t.Fatal(err)
	}
	p := eg.Profiles["a@test.com"]
	if p.DeviceMode != QoderDeviceModeAuto {
		t.Errorf("device_mode 应为 auto，实际 %q", p.DeviceMode)
	}
	if p.ProxyURL != "http://127.0.0.1:1080" {
		t.Errorf("代理应保留，实际 %q", p.ProxyURL)
	}
	if _, err := os.Stat(egPath); err != nil {
		t.Fatalf("出口配置文件应存在: %v", err)
	}
}
