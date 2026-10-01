package lod

import (
	"math/rand"
	"testing"
)

func TestNearIsLevelZero(t *testing.T) {
	manager := NewManager()
	manager.Add("e1", 10)
	if level := manager.Tick("e1", 10); level != 0 {
		t.Fatalf("近距离应该是 0 级：%d", level)
	}
}

func TestFarIsLevelTwo(t *testing.T) {
	manager := NewManager()
	manager.Add("e1", 10)
	if level := manager.Tick("e1", 100); level != 2 {
		t.Fatalf("远距离应该是 2 级：%d", level)
	}
}

func TestUnknownEntityIsIgnored(t *testing.T) {
	if level := NewManager().Tick("missing", 10); level != -1 {
		t.Fatalf("不存在的实体应该返回 -1：%d", level)
	}
}

func TestConstants(t *testing.T) {
	if NearDistance != 20 || FarDistance != 60 || Hysteresis != 5 {
		t.Fatal("LOD 常量被改了")
	}
}

// 阈值含边界：distance == NearDistance 是 0 级，distance == FarDistance 是 1 级。
func TestBoundaryLevels(t *testing.T) {
	cases := []struct {
		distance float64
		level    int
	}{
		{NearDistance, 0},
		{NearDistance - 0.001, 0},
		{NearDistance + 0.001, 1},
		{FarDistance, 1},
		{FarDistance - 0.001, 1},
		{FarDistance + 0.001, 2},
	}
	for _, tc := range cases {
		manager := NewManager()
		manager.Add("e1", tc.distance)
		if got := manager.Level("e1"); got != tc.level {
			t.Fatalf("距离 %g 的初始级别应该是 %d：%d", tc.distance, tc.level, got)
		}
	}
}

// 站在近阈值边界上（20）仍判 0 级，不会因边界差一跳到 1 级。
func TestTickOnBoundaryStaysLevelZero(t *testing.T) {
	manager := NewManager()
	manager.Add("e1", 10)
	if level := manager.Tick("e1", NearDistance); level != 0 {
		t.Fatalf("距离正好 %g 应该是 0 级：%d", NearDistance, level)
	}
	if manager.Switches() != 0 {
		t.Fatalf("边界上不应产生切换：%d", manager.Switches())
	}
}

// 滞回带里来回抖不切换；只有大出/小出 Hysteresis 才切换。
func TestHysteresisBand(t *testing.T) {
	// 从 0 级出发，22 与 18 都在近阈值滞回带 (15,25) 内，级别恒为 0。
	manager := NewManager()
	manager.Add("e1", 10)
	if manager.Tick("e1", 10) != 0 || manager.Switches() != 0 {
		t.Fatal("初始推进不应记切换")
	}
	start := manager.Switches()
	for index := 0; index < 10; index++ {
		if level := manager.Tick("e1", 22); level != 0 {
			t.Fatalf("滞回带里 22 不应升级：%d", level)
		}
		if level := manager.Tick("e1", 18); level != 0 {
			t.Fatalf("滞回带里 18 不应升级：%d", level)
		}
	}
	if got := manager.Switches() - start; got != 0 {
		t.Fatalf("滞回带内抖动不应产生切换：%d", got)
	}

	// 严格大于 NearDistance+Hysteresis 才升级到 1；之后回到阈值本身仍保持 1 级。
	if level := manager.Tick("e1", NearDistance+Hysteresis); level != 0 {
		t.Fatalf("距离等于阈值+滞回宽度时不应升级：%d", level)
	}
	if level := manager.Tick("e1", NearDistance+Hysteresis+0.001); level != 1 {
		t.Fatalf("超出滞回带后应升到 1 级：%d", level)
	}
	if level := manager.Tick("e1", NearDistance); level != 1 {
		t.Fatalf("回到阈值上应保持 1 级（滞回）：%d", level)
	}
	if level := manager.Tick("e1", NearDistance-Hysteresis); level != 1 {
		t.Fatalf("距离等于阈值-滞回宽度时不应降级：%d", level)
	}
	if level := manager.Tick("e1", NearDistance-Hysteresis-0.001); level != 0 {
		t.Fatalf("小出滞回带后应降回 0 级：%d", level)
	}

	// 从 1 级出发在 22/18 抖动同样不切换。
	switches := manager.Switches()
	for index := 0; index < 5; index++ {
		manager.Tick("e1", 25.001)
		manager.Tick("e1", 18)
		manager.Tick("e1", 22)
	}
	if manager.Level("e1") != 1 {
		t.Fatalf("抖动结束后应停在 1 级：%d", manager.Level("e1"))
	}
	if got := manager.Switches() - switches; got != 1 {
		t.Fatalf("这段抖动应只有 0→1 一次切换：%d", got)
	}
}

// 远阈值滞回：1→2 要大于 65，2→1 要小于 55。
func TestFarThresholdHysteresis(t *testing.T) {
	manager := NewManager()
	manager.Add("e1", 40)
	if level := manager.Tick("e1", FarDistance+Hysteresis); level != 1 {
		t.Fatalf("距离等于远阈值+滞回宽度时不应升级：%d", level)
	}
	if level := manager.Tick("e1", FarDistance+Hysteresis+0.001); level != 2 {
		t.Fatalf("超出远阈值滞回带应升到 2 级：%d", level)
	}
	for index := 0; index < 10; index++ {
		if level := manager.Tick("e1", 56); level != 2 {
			t.Fatalf("远阈值滞回带里 56 不应降级：%d", level)
		}
		if level := manager.Tick("e1", 64); level != 2 {
			t.Fatalf("远阈值滞回带里 64 不应降级：%d", level)
		}
	}
	if level := manager.Tick("e1", FarDistance-Hysteresis-0.001); level != 1 {
		t.Fatalf("小出远阈值滞回带应降回 1 级：%d", level)
	}
}

// 距离不变反复推进不累加 Switches。
func TestNoSwitchWhenLevelUnchanged(t *testing.T) {
	manager := NewManager()
	manager.Add("e1", 10)
	for index := 0; index < 10; index++ {
		if level := manager.Tick("e1", 10); level != 0 {
			t.Fatalf("同距离推进级别应保持 0：%d", level)
		}
	}
	if manager.Switches() != 0 {
		t.Fatalf("同级别重复推进不应记切换：%d", manager.Switches())
	}
}

// 真实变化时 Switches 才累加，且累计值与变化次数一致。
func TestSwitchCountMatchesChanges(t *testing.T) {
	manager := NewManager()
	manager.Add("e1", 10)
	steps := []struct {
		distance float64
		level    int
	}{
		{25.001, 1},
		{25.001, 1},
		{40, 1},
		{65.001, 2},
		{100, 2},
		{54.999, 1},
		{14.999, 0},
		{10, 0},
	}
	for index, step := range steps {
		if level := manager.Tick("e1", step.distance); level != step.level {
			t.Fatalf("第 %d 步距离 %g 级别应为 %d：%d", index, step.distance, step.level, level)
		}
	}
	if manager.Switches() != 4 {
		t.Fatalf("0→1→2→1→0 应有 4 次切换：%d", manager.Switches())
	}
}

// 查询代价：Tick 只碰目标实体，Scanned 随推进次数线性、不随实体数线性。
func TestTickScanCostIsConstant(t *testing.T) {
	manager := NewManager()
	const entityCount = 3000
	for index := 0; index < entityCount; index++ {
		manager.Add(idFor(index), 10)
	}
	before := manager.Scanned()
	const ticks = 3000
	for index := 0; index < ticks; index++ {
		manager.Tick(idFor(entityCount-1), 10)
	}
	got := manager.Scanned() - before
	if got != ticks {
		t.Fatalf("%d 次推进的扫描量应是 %d（每次 O(1)）：%d", ticks, ticks, got)
	}
	if got > entityCount {
		t.Fatalf("扫描量 %d 不应随实体数 %d 线性增长", got, entityCount)
	}
}

func idFor(index int) string {
	return "e-" + itoa(index)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[pos:])
}

// boundaryLevel 是参考实现的含边界定级。
func boundaryLevel(distance float64) int {
	if distance <= NearDistance {
		return 0
	}
	if distance <= FarDistance {
		return 1
	}
	return 2
}

// thresholdBetween 返回 level-1 与 level 两级之间的阈值（1 用 NearDistance，2 用 FarDistance）。
func thresholdBetween(level int) float64 {
	if level == 1 {
		return NearDistance
	}
	return FarDistance
}

// referenceStep 是参考实现：含边界定级 + 升级大出滞回、降级小出滞回、滞回带内不动。
// 写法刻意与 Manager.Tick 不同（逐级试探跨阈值），用于交叉自证。
func referenceStep(level int, distance float64) int {
	for level < 2 && distance > thresholdBetween(level+1)+Hysteresis {
		level++
	}
	for level > 0 && distance < thresholdBetween(level)-Hysteresis {
		level--
	}
	return level
}

// 随机 200 组距离序列（含边界与抖动）与参考实现逐次比对级别与切换次数。
func TestRandomSequencesAgainstReference(t *testing.T) {
	random := rand.New(rand.NewSource(20261002))
	special := []float64{
		0, 10, 14.999, 15, 18, 20, 22, 25, 25.001,
		40, 54.999, 55, 60, 62, 65, 65.001, 100,
		NearDistance, FarDistance,
		NearDistance - Hysteresis, NearDistance + Hysteresis,
		FarDistance - Hysteresis, FarDistance + Hysteresis,
	}
	for trial := 0; trial < 200; trial++ {
		length := random.Intn(60) + 1
		distances := make([]float64, length)
		for index := range distances {
			if random.Intn(2) == 0 {
				distances[index] = special[random.Intn(len(special))]
			} else {
				// 以某个边界为中心做小范围抖动。
				center := special[random.Intn(len(special))]
				distances[index] = center + (random.Float64()-0.5)*2
			}
		}

		// 参考实现（实体初始级别与 Add 一样按含边界口径定级）。
		refLevel := boundaryLevel(10)
		refSwitches := 0
		manager := NewManager()
		manager.Add("e1", 10)
		replay := NewManager()
		replay.Add("e1", 10)
		firstLevels := make([]int, 0, length)

		for index, distance := range distances {
			next := referenceStep(refLevel, distance)
			if next != refLevel {
				refSwitches++
			}
			refLevel = next

			got := manager.Tick("e1", distance)
			if got != refLevel {
				t.Fatalf("trial=%d step=%d distance=%g：级别 %d 与参考 %d 不一致",
					trial, index, distance, got, refLevel)
			}
			firstLevels = append(firstLevels, got)

			if level := replay.Tick("e1", distance); level != got {
				t.Fatalf("trial=%d step=%d：同一序列重复执行级别不一致：%d != %d",
					trial, index, level, got)
			}
		}
		if manager.Switches() != refSwitches {
			t.Fatalf("trial=%d：切换次数 %d 与参考 %d 不一致", trial, manager.Switches(), refSwitches)
		}
		if replay.Switches() != manager.Switches() {
			t.Fatalf("trial=%d：同一序列重复执行切换次数不一致：%d != %d",
				trial, replay.Switches(), manager.Switches())
		}
	}
}
