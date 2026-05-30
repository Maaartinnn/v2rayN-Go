# Progress

## Current Status

策略组重构：StrategyGroup 表 → Profile 虚拟节点（组合模式）

### 已完成
- 项目基础架构搭建
- 前端精简传输重构（ProfileListItem DTO）
- 通用 Toast 通知系统
- 后端代码质量提升（常量提取、slog、错误链）
- 测试编写（19 个 Go 测试 + 3 个前端测试）
- CI/CD 配置
- 断电安全防护改造（原子写入 + .bak 容灾回滚 + SQLite WAL 模式）
- 无文件落地（Fileless Execution）：stdin 管道注入 + 跨平台进程安全 + 调试开关

### 进行中
- 策略组重构：将 StrategyGroup 合并进 Profile 表

### 待完成
- 实施重构（见下方详细清单）

## Known Issues
- 无
