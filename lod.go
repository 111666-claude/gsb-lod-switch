// Package lod 是实体细节层级：按距离切换 LOD、用滞回防抖、切换事件只在真正变化时产生。
// 缺陷：没有滞回、每次调用都记一次切换、阈值边界差一、每次查询全扫实体表。
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

// Add 加一个实体。
func (m *Manager) Add(id string, distance float64) {
	m.entities[id] = &Entity{ID: id, Distance: distance}
}

// Tick 更新距离并返回 LOD 级别。
// 缺陷：每次调用都累加切换次数（没变化也算）、阈值用 >= 判边界、没有滞回、每次全扫实体表。
func (m *Manager) Tick(id string, distance float64) int {
	m.scanned += len(m.entities)
	entity := m.entities[id]
	if entity == nil {
		return -1
	}
	entity.Distance = distance
	level := 0
	if distance >= NearDistance {
		level = 1
	}
	if distance >= FarDistance {
		level = 2
	}
	entity.Level = level
	m.switches++
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
