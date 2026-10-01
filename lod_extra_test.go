package lod

import (
	"math/rand"
	"testing"
)

// 参考实现：阈值含边界 + 滞回 + 变化才计数。
type reference struct {
	level    int
	switches int
}

func newReference(distance float64) *reference {
	return &reference{level: referenceBase(distance)}
}

func referenceBase(distance float64) int {
	switch {
	case distance <= NearDistance:
		return 0
	case distance <= FarDistance:
		return 1
	default:
		return 2
	}
}

func (r *reference) tick(distance float64) int {
	next := r.level
	switch r.level {
	case 0:
		if distance > FarDistance+Hysteresis {
			next = 2
		} else if distance > NearDistance+Hysteresis {
			next = 1
		}
	case 1:
		if distance > FarDistance+Hysteresis {
			next = 2
		} else if distance < NearDistance-Hysteresis {
			next = 0
		}
	case 2:
		if distance < NearDistance-Hysteresis {
			next = 0
		} else if distance < FarDistance-Hysteresis {
			next = 1
		}
	}
	if next != r.level {
		r.level = next
		r.switches++
	}
	return r.level
}

func TestBoundaryLevels(t *testing.T) {
	cases := []struct {
		distance float64
		want     int
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
		if level := manager.Level("e1"); level != tc.want {
			t.Fatalf("distance=%v 应为 %d 级，得到 %d", tc.distance, tc.want, level)
		}
	}
}

func TestHysteresisDebounce(t *testing.T) {
	manager := NewManager()
	manager.Add("e1", 10)
	manager.Tick("e1", 10)
	start := manager.Switches()
	for index := 0; index < 10; index++ {
		if level := manager.Tick("e1", 22); level != 0 {
			t.Fatalf("滞回带内应保持 0 级，得到 %d", level)
		}
		if level := manager.Tick("e1", 18); level != 0 {
			t.Fatalf("滞回带内应保持 0 级，得到 %d", level)
		}
	}
	if got := manager.Switches() - start; got != 0 {
		t.Fatalf("滞回带内抖动不应记切换，得到 %d", got)
	}
}

func TestHysteresisCrossing(t *testing.T) {
	manager := NewManager()
	manager.Add("e1", 10)
	manager.Tick("e1", 10)
	if level := manager.Tick("e1", NearDistance+Hysteresis); level != 0 {
		t.Fatalf("恰好到升级沿不应升级，得到 %d", level)
	}
	if level := manager.Tick("e1", NearDistance+Hysteresis+0.001); level != 1 {
		t.Fatalf("越过升级沿应升 1 级，得到 %d", level)
	}
	if level := manager.Tick("e1", NearDistance-Hysteresis); level != 1 {
		t.Fatalf("恰好到降级沿不应降级，得到 %d", level)
	}
	if level := manager.Tick("e1", NearDistance-Hysteresis-0.001); level != 0 {
		t.Fatalf("越过降级沿应降 0 级，得到 %d", level)
	}
	if got := manager.Switches(); got != 2 {
		t.Fatalf("两次真正切换应为 2，得到 %d", got)
	}
}

func TestSwitchCountOnlyOnChange(t *testing.T) {
	manager := NewManager()
	manager.Add("e1", 10)
	for index := 0; index < 10; index++ {
		manager.Tick("e1", 10)
	}
	if got := manager.Switches(); got != 0 {
		t.Fatalf("同级别重复推进不应计数，得到 %d", got)
	}
	manager.Tick("e1", 100)
	manager.Tick("e1", 100)
	manager.Tick("e1", 100)
	if got := manager.Switches(); got != 1 {
		t.Fatalf("同级别重复推进不应计数，得到 %d", got)
	}
}

func TestQueryCostIsConstant(t *testing.T) {
	manager := NewManager()
	for index := 0; index < 50000; index++ {
		manager.Add(string(rune(index)), 10)
	}
	for index := 0; index < 3000; index++ {
		manager.Tick(string(rune(0)), 10)
	}
	if got := manager.Scanned(); got > 3000 {
		t.Fatalf("Scanned 不应随实体数线性增长，得到 %d", got)
	}
}

func TestDeterministicReplay(t *testing.T) {
	sequence := []float64{10, 22, 18, 22, 26, 70, 54, 20, 60, 61, 14, 15, 16}
	run := func() (int, int) {
		manager := NewManager()
		manager.Add("e1", 10)
		level := 0
		for _, distance := range sequence {
			level = manager.Tick("e1", distance)
		}
		return level, manager.Switches()
	}
	levelA, switchesA := run()
	levelB, switchesB := run()
	if levelA != levelB || switchesA != switchesB {
		t.Fatalf("同一序列重复执行结果不同：(%d,%d) vs (%d,%d)", levelA, switchesA, levelB, switchesB)
	}
}

func TestAgainstReference(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	interesting := []float64{
		0, 10,
		NearDistance - Hysteresis, NearDistance, NearDistance + Hysteresis,
		FarDistance - Hysteresis, FarDistance, FarDistance + Hysteresis,
		18, 22, 100,
	}
	for trial := 0; trial < 200; trial++ {
		manager := NewManager()
		manager.Add("e1", 10)
		manager.Tick("e1", 10)
		ref := newReference(10)
		steps := 1 + rng.Intn(50)
		for step := 0; step < steps; step++ {
			var distance float64
			if rng.Intn(2) == 0 {
				distance = interesting[rng.Intn(len(interesting))]
			} else {
				distance = rng.Float64() * 120
			}
			if rng.Intn(3) == 0 {
				distance += rng.Float64()*2 - 1 // 边界附近抖动
			}
			gotLevel := manager.Tick("e1", distance)
			wantLevel := ref.tick(distance)
			if gotLevel != wantLevel {
				t.Fatalf("trial=%d step=%d distance=%v 级别不一致：got=%d want=%d",
					trial, step, distance, gotLevel, wantLevel)
			}
			if manager.Switches() != ref.switches {
				t.Fatalf("trial=%d step=%d distance=%v 切换数不一致：got=%d want=%d",
					trial, step, distance, manager.Switches(), ref.switches)
			}
		}
	}
}
