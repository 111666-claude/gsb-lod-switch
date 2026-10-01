package lod

import "testing"

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
