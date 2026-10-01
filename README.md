# lod-switch

实体细节层级：按距离切换 LOD、用滞回防抖、切换事件只在真正变化时产生，只用 Go 标准库。

```
go test ./...
go run ./cmd/lod-switch --sample hysteresis
go run ./cmd/lod-switch --sample event
go run ./cmd/lod-switch --sample boundary
```

## 口径（README 为准）

- **级别含边界**：`distance <= NearDistance` 是 0 级，`<= FarDistance` 是 1 级，其余 2 级。
- **滞回防抖**：升级要比阈值大出 `Hysteresis`，降级要比阈值小出 `Hysteresis`，
  距离在滞回带里来回抖动时级别保持不变。
- **切换事件只在变化时产生**：同一级别重复推进会话不累加 `Switches`。
- **查询代价**：`Tick` 只碰目标实体，`Scanned` 不随实体数线性增长。
- **规模**：单场 5 万实体、每秒 60 次推进，单次 O(1)，内存 O(实体数)。

## 输出契约（不改格式）

```
switches=0
switches=0
level=0
scanned<=3000
```

## 常量

```
NearDistance = 20
FarDistance  = 60
Hysteresis   = 5
```

## 目录

```
lod.go                  LOD 切换
cmd/lod-switch          命令行入口
lod_test.go             go test 用例
```
