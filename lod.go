// Package lod 是实体细节层级：按距离切换 LOD、用滞回防抖、切换事件只在真正变化时产生。
package lod

// 距离阈值与滞回宽度。
const (
	NearDistance = 20.0
	FarDistance  = 60.0
	Hysteresis   = 5.0
)

// Entity 是一个实体。
type Entity struct {
	ID       string
	Distance float64
	Level    int
}

// Manager 是 LOD 管理器。
type Manager struct {
	entities map[string]*Entity
	switches int
	scanned  int
}

// NewManager 建管理器。
func NewManager() *Manager {
	return &Manager{entities: map[string]*Entity{}}
}

// Add 加一个实体，初始级别按含边界阈值直接判定。
func (m *Manager) Add(id string, distance float64) {
	m.entities[id] = &Entity{ID: id, Distance: distance, Level: baseLevel(distance)}
}

// baseLevel 按含边界口径判定级别：
// distance <= NearDistance 是 0 级，<= FarDistance 是 1 级，其余 2 级。
func baseLevel(distance float64) int {
	switch {
	case distance <= NearDistance:
		return 0
	case distance <= FarDistance:
		return 1
	default:
		return 2
	}
}

// Tick 更新距离并返回 LOD 级别。
// 升级要比阈值大出 Hysteresis，降级要比阈值小出 Hysteresis，滞回带内级别不变；
// 只碰目标实体（O(1)），且仅在级别真正变化时累加 Switches。
func (m *Manager) Tick(id string, distance float64) int {
	entity := m.entities[id]
	if entity == nil {
		return -1
	}
	m.scanned++
	entity.Distance = distance
	next := stepLevel(entity.Level, distance)
	if next != entity.Level {
		entity.Level = next
		m.switches++
	}
	return entity.Level
}

// stepLevel 在当前级别上应用含滞回的转移规则。
func stepLevel(level int, distance float64) int {
	switch level {
	case 0:
		if distance > FarDistance+Hysteresis {
			return 2
		}
		if distance > NearDistance+Hysteresis {
			return 1
		}
	case 1:
		if distance > FarDistance+Hysteresis {
			return 2
		}
		if distance < NearDistance-Hysteresis {
			return 0
		}
	case 2:
		if distance < NearDistance-Hysteresis {
			return 0
		}
		if distance < FarDistance-Hysteresis {
			return 1
		}
	}
	return level
}

// Level 是当前级别。
func (m *Manager) Level(id string) int {
	if entity := m.entities[id]; entity != nil {
		return entity.Level
	}
	return -1
}

// Switches 是累计切换次数。
func (m *Manager) Switches() int { return m.switches }

// Scanned 是累计扫描的实体数（规模观测）。
func (m *Manager) Scanned() int { return m.scanned }
