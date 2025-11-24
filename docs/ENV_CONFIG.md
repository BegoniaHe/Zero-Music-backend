# 环境变量配置说明

Zero-Music 后端支持通过环境变量覆盖配置文件中的设置。环境变量的优先级高于配置文件。

> **最后更新**: 2025-11-23  
> **版本**: v0.5

## 可用的环境变量

### 服务器配置

| 环境变量 | 说明 | 默认值 | 有效范围 | 示例 |
|---------|------|--------|---------|------|
| `ZERO_MUSIC_SERVER_HOST` | 服务器监听地址 | `0.0.0.0` | 任意有效 IP 地址 | `ZERO_MUSIC_SERVER_HOST=127.0.0.1` |
| `ZERO_MUSIC_SERVER_PORT` | 服务器监听端口 | `8080` | `1-65535` | `ZERO_MUSIC_SERVER_PORT=3000` |
| `ZERO_MUSIC_MAX_RANGE_SIZE` | 单次 Range 请求最大字节数 | `104857600` (100MB) | `1-524288000` (500MB) | `ZERO_MUSIC_MAX_RANGE_SIZE=52428800` |
| `ZERO_MUSIC_SERVER_READ_TIMEOUT_SECONDS` | HTTP 读取超时（秒） | `15` | `1-600` | `ZERO_MUSIC_SERVER_READ_TIMEOUT_SECONDS=30` |
| `ZERO_MUSIC_SERVER_WRITE_TIMEOUT_SECONDS` | HTTP 写入超时（秒） | `60` | `1-600` | `ZERO_MUSIC_SERVER_WRITE_TIMEOUT_SECONDS=120` |
| `ZERO_MUSIC_SERVER_IDLE_TIMEOUT_SECONDS` | HTTP 空闲连接超时（秒） | `120` | `1-600` | `ZERO_MUSIC_SERVER_IDLE_TIMEOUT_SECONDS=180` |
| `ZERO_MUSIC_SERVER_SHUTDOWN_TIMEOUT_SECONDS` | 服务器优雅关闭超时（秒） | `30` | `1-300` | `ZERO_MUSIC_SERVER_SHUTDOWN_TIMEOUT_SECONDS=60` |

### 音乐库配置

| 环境变量 | 说明 | 默认值 | 有效范围 | 示例 |
|---------|------|--------|---------|------|
| `ZERO_MUSIC_MUSIC_DIRECTORY` | 音乐文件目录 | `~/Music` 或 `./music` | 任意存在的目录路径 | `ZERO_MUSIC_MUSIC_DIRECTORY=/data/music` |
| `ZERO_MUSIC_CACHE_TTL_MINUTES` | 缓存有效期（分钟） | `5` | `1-1440` (24小时) | `ZERO_MUSIC_CACHE_TTL_MINUTES=10` |

### 调试与日志配置

| 环境变量 | 说明 | 默认值 | 有效值 | 示例 |
|---------|------|--------|-------|------|
| `ZERO_MUSIC_DEBUG` | 是否启用调试模式（显示详细错误） | `false` | `true` / `false` | `ZERO_MUSIC_DEBUG=true` |
| `LOG_LEVEL` | 日志级别 | `info` | `debug`, `info`, `warn`, `error`, `fatal` | `LOG_LEVEL=debug` |

## 使用方法

### 方法一：直接设置环境变量

#### Linux/macOS

```bash
# 基础配置
export ZERO_MUSIC_SERVER_HOST=0.0.0.0
export ZERO_MUSIC_SERVER_PORT=8080
export ZERO_MUSIC_MUSIC_DIRECTORY=/path/to/music

# 超时配置（可选）
export ZERO_MUSIC_SERVER_READ_TIMEOUT_SECONDS=15
export ZERO_MUSIC_SERVER_WRITE_TIMEOUT_SECONDS=60
export ZERO_MUSIC_SERVER_IDLE_TIMEOUT_SECONDS=120
export ZERO_MUSIC_SERVER_SHUTDOWN_TIMEOUT_SECONDS=30

# 性能配置（可选）
export ZERO_MUSIC_MAX_RANGE_SIZE=104857600  # 100MB
export ZERO_MUSIC_CACHE_TTL_MINUTES=5

# 调试配置（开发环境）
export ZERO_MUSIC_DEBUG=false  # 生产环境设为 false
export LOG_LEVEL=info

# 启动服务
./zero-music -config config.json -log app.log
```

#### Windows (PowerShell)

```powershell
# 基础配置
$env:ZERO_MUSIC_SERVER_HOST="0.0.0.0"
$env:ZERO_MUSIC_SERVER_PORT=8080
$env:ZERO_MUSIC_MUSIC_DIRECTORY="C:\Music"

# 超时配置（可选）
$env:ZERO_MUSIC_SERVER_READ_TIMEOUT_SECONDS=15
$env:ZERO_MUSIC_SERVER_WRITE_TIMEOUT_SECONDS=60

# 调试配置
$env:ZERO_MUSIC_DEBUG="false"
$env:LOG_LEVEL="info"

# 启动服务
.\zero-music.exe -config config.json -log app.log
```

#### Windows (CMD)

```cmd
REM 基础配置
set ZERO_MUSIC_SERVER_PORT=8080
set ZERO_MUSIC_MUSIC_DIRECTORY=C:\Music
set ZERO_MUSIC_DEBUG=false

REM 启动服务
zero-music.exe -config config.json -log app.log
```

### 方法二：使用配置文件（推荐生产环境）

创建一个 shell 脚本来管理环境变量：

#### production.sh (生产环境)

```bash
#!/bin/bash

# 服务器配置
export ZERO_MUSIC_SERVER_HOST=0.0.0.0
export ZERO_MUSIC_SERVER_PORT=8080

# 超时配置
export ZERO_MUSIC_SERVER_READ_TIMEOUT_SECONDS=15
export ZERO_MUSIC_SERVER_WRITE_TIMEOUT_SECONDS=60
export ZERO_MUSIC_SERVER_IDLE_TIMEOUT_SECONDS=120
export ZERO_MUSIC_SERVER_SHUTDOWN_TIMEOUT_SECONDS=30

# 性能配置
export ZERO_MUSIC_MAX_RANGE_SIZE=104857600  # 100MB
export ZERO_MUSIC_CACHE_TTL_MINUTES=5

# 音乐目录
export ZERO_MUSIC_MUSIC_DIRECTORY=/data/music

# 安全配置 - 生产环境必须设置
export ZERO_MUSIC_DEBUG=false  # ⚠️ 禁用调试模式
export LOG_LEVEL=info

echo "生产环境配置已加载"
echo "服务地址: $ZERO_MUSIC_SERVER_HOST:$ZERO_MUSIC_SERVER_PORT"
echo "音乐目录: $ZERO_MUSIC_MUSIC_DIRECTORY"
echo "调试模式: $ZERO_MUSIC_DEBUG"
```

#### development.sh (开发环境)

```bash
#!/bin/bash

# 服务器配置
export ZERO_MUSIC_SERVER_HOST=127.0.0.1
export ZERO_MUSIC_SERVER_PORT=8080

# 超时配置（开发环境可以更宽松）
export ZERO_MUSIC_SERVER_READ_TIMEOUT_SECONDS=30
export ZERO_MUSIC_SERVER_WRITE_TIMEOUT_SECONDS=120

# 音乐目录
export ZERO_MUSIC_MUSIC_DIRECTORY=./music

# 调试配置 - 开发环境启用详细日志
export ZERO_MUSIC_DEBUG=true  # 启用详细错误信息
export LOG_LEVEL=debug

echo "开发环境配置已加载"
echo "服务地址: $ZERO_MUSIC_SERVER_HOST:$ZERO_MUSIC_SERVER_PORT"
echo "音乐目录: $ZERO_MUSIC_MUSIC_DIRECTORY"
echo "调试模式: $ZERO_MUSIC_DEBUG"
```

#### 使用方法

```bash
# 生产环境
source production.sh
./zero-music -config config.json -log /var/log/zero-music/app.log

# 开发环境
source development.sh
./zero-music -config config.json -log app.log
```

### 方法三：使用 .env 文件（配合第三方工具）

1. 创建 `.env` 文件：

```bash
# .env 文件示例
ZERO_MUSIC_SERVER_HOST=0.0.0.0
ZERO_MUSIC_SERVER_PORT=8080
ZERO_MUSIC_MUSIC_DIRECTORY=/data/music

# 超时配置
ZERO_MUSIC_SERVER_READ_TIMEOUT_SECONDS=15
ZERO_MUSIC_SERVER_WRITE_TIMEOUT_SECONDS=60
ZERO_MUSIC_SERVER_IDLE_TIMEOUT_SECONDS=120
ZERO_MUSIC_SERVER_SHUTDOWN_TIMEOUT_SECONDS=30

# 性能配置
ZERO_MUSIC_MAX_RANGE_SIZE=104857600
ZERO_MUSIC_CACHE_TTL_MINUTES=5

# 安全配置
ZERO_MUSIC_DEBUG=false
LOG_LEVEL=info
```

2. 使用 `godotenv` 工具加载：

```bash
# 安装 godotenv
go install github.com/joho/godotenv/cmd/godotenv@latest

# 运行
godotenv -f .env ./zero-music -config config.json -log app.log
```

3. 或者使用 `direnv`（自动加载）：

```bash
# 安装 direnv (macOS)
brew install direnv

# 配置 shell (添加到 ~/.bashrc 或 ~/.zshrc)
eval "$(direnv hook bash)"  # 或 eval "$(direnv hook zsh)"

# 创建 .envrc 文件
cp .env .envrc

# 允许 direnv 加载配置
direnv allow

# 进入目录时自动加载环境变量
cd /path/to/zero-music-backend
# direnv: loading .envrc
```

### 方法四：Docker 环境

#### Docker Run

```bash
docker run -d \
  --name zero-music \
  -e ZERO_MUSIC_SERVER_HOST=0.0.0.0 \
  -e ZERO_MUSIC_SERVER_PORT=8080 \
  -e ZERO_MUSIC_SERVER_READ_TIMEOUT_SECONDS=15 \
  -e ZERO_MUSIC_SERVER_WRITE_TIMEOUT_SECONDS=60 \
  -e ZERO_MUSIC_SERVER_IDLE_TIMEOUT_SECONDS=120 \
  -e ZERO_MUSIC_SERVER_SHUTDOWN_TIMEOUT_SECONDS=30 \
  -e ZERO_MUSIC_MUSIC_DIRECTORY=/music \
  -e ZERO_MUSIC_MAX_RANGE_SIZE=104857600 \
  -e ZERO_MUSIC_CACHE_TTL_MINUTES=5 \
  -e ZERO_MUSIC_DEBUG=false \
  -e LOG_LEVEL=info \
  -v /path/to/music:/music:ro \
  -v /path/to/logs:/logs \
  -p 8080:8080 \
  zero-music:latest
```

#### Docker Compose

创建 `docker-compose.yml` 文件：

```yaml
version: '3.8'

services:
  zero-music:
    image: zero-music:latest
    container_name: zero-music
    environment:
      # 服务器配置
      ZERO_MUSIC_SERVER_HOST: 0.0.0.0
      ZERO_MUSIC_SERVER_PORT: 8080
      
      # 超时配置
      ZERO_MUSIC_SERVER_READ_TIMEOUT_SECONDS: 15
      ZERO_MUSIC_SERVER_WRITE_TIMEOUT_SECONDS: 60
      ZERO_MUSIC_SERVER_IDLE_TIMEOUT_SECONDS: 120
      ZERO_MUSIC_SERVER_SHUTDOWN_TIMEOUT_SECONDS: 30
      
      # 性能配置
      ZERO_MUSIC_MAX_RANGE_SIZE: 104857600  # 100MB
      ZERO_MUSIC_CACHE_TTL_MINUTES: 5
      
      # 音乐目录
      ZERO_MUSIC_MUSIC_DIRECTORY: /music
      
      # 安全配置
      ZERO_MUSIC_DEBUG: "false"
      LOG_LEVEL: info
    
    volumes:
      - /path/to/music:/music:ro  # 只读挂载音乐目录
      - /path/to/logs:/logs       # 日志目录
    
    ports:
      - "8080:8080"
    
    restart: unless-stopped
    
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
```

使用 docker-compose：

```bash
# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

#### Docker Compose with .env 文件

创建 `.env` 文件：

```bash
# .env
ZERO_MUSIC_SERVER_PORT=8080
ZERO_MUSIC_MUSIC_DIRECTORY=/music
ZERO_MUSIC_DEBUG=false
LOG_LEVEL=info
MUSIC_HOST_PATH=/data/music
LOGS_HOST_PATH=/var/log/zero-music
```

修改 `docker-compose.yml`：

```yaml
version: '3.8'

services:
  zero-music:
    image: zero-music:latest
    environment:
      ZERO_MUSIC_SERVER_PORT: ${ZERO_MUSIC_SERVER_PORT:-8080}
      ZERO_MUSIC_MUSIC_DIRECTORY: ${ZERO_MUSIC_MUSIC_DIRECTORY:-/music}
      ZERO_MUSIC_DEBUG: ${ZERO_MUSIC_DEBUG:-false}
      LOG_LEVEL: ${LOG_LEVEL:-info}
    volumes:
      - ${MUSIC_HOST_PATH}:/music:ro
      - ${LOGS_HOST_PATH}:/logs
    ports:
      - "${ZERO_MUSIC_SERVER_PORT}:${ZERO_MUSIC_SERVER_PORT}"
```

### 方法五：Systemd 服务（Linux）

创建 systemd 服务文件 `/etc/systemd/system/zero-music.service`：

```ini
[Unit]
Description=Zero Music Backend Service
After=network.target

[Service]
Type=simple
User=zeromusic
Group=zeromusic
WorkingDirectory=/opt/zero-music

# 环境变量配置
Environment="ZERO_MUSIC_SERVER_HOST=0.0.0.0"
Environment="ZERO_MUSIC_SERVER_PORT=8080"
Environment="ZERO_MUSIC_SERVER_READ_TIMEOUT_SECONDS=15"
Environment="ZERO_MUSIC_SERVER_WRITE_TIMEOUT_SECONDS=60"
Environment="ZERO_MUSIC_SERVER_IDLE_TIMEOUT_SECONDS=120"
Environment="ZERO_MUSIC_SERVER_SHUTDOWN_TIMEOUT_SECONDS=30"
Environment="ZERO_MUSIC_MUSIC_DIRECTORY=/data/music"
Environment="ZERO_MUSIC_MAX_RANGE_SIZE=104857600"
Environment="ZERO_MUSIC_CACHE_TTL_MINUTES=5"
Environment="ZERO_MUSIC_DEBUG=false"
Environment="LOG_LEVEL=info"

# 或者从文件加载环境变量
EnvironmentFile=-/etc/zero-music/environment

ExecStart=/opt/zero-music/zero-music -config /etc/zero-music/config.json -log /var/log/zero-music/app.log

Restart=always
RestartSec=10

# 安全加固
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/log/zero-music

[Install]
WantedBy=multi-user.target
```

创建环境变量文件 `/etc/zero-music/environment`：

```bash
ZERO_MUSIC_SERVER_HOST=0.0.0.0
ZERO_MUSIC_SERVER_PORT=8080
ZERO_MUSIC_MUSIC_DIRECTORY=/data/music
ZERO_MUSIC_DEBUG=false
LOG_LEVEL=info
```

管理服务：

```bash
# 重新加载 systemd 配置
sudo systemctl daemon-reload

# 启动服务
sudo systemctl start zero-music

# 设置开机自启
sudo systemctl enable zero-music

# 查看状态
sudo systemctl status zero-music

# 查看日志
sudo journalctl -u zero-music -f
```

## 配置优先级

配置的加载优先级从高到低为：

1. **环境变量** ← 最高优先级
2. **配置文件** (`config.json`)
3. **默认值** ← 最低优先级

### 示例

假设有以下配置：

**config.json**:

```json
{
  "server": {
    "port": 8080,
    "read_timeout_seconds": 10
  }
}
```

**环境变量**:

```bash
export ZERO_MUSIC_SERVER_PORT=3000
```

**最终生效的配置**:

- 端口: `3000` （来自环境变量）
- 读取超时: `10` （来自配置文件）
- 写入超时: `60` （来自默认值）

## 配置验证

### 启动日志

服务器启动时会显示实际使用的配置：

```json
{
  "level": "info",
  "msg": "Zero Music 服务器启动中...",
  "time": "2025-11-23 18:30:00"
}
{
  "level": "info",
  "msg": "服务地址: http://localhost:8080",
  "time": "2025-11-23 18:30:00"
}
{
  "level": "info",
  "msg": "音乐目录: /data/music",
  "time": "2025-11-23 18:30:00"
}
```

### 健康检查

使用健康检查端点验证配置：

```bash
curl http://localhost:8080/health
```

响应示例：

```json
{
  "status": "ok",
  "message": "zero music服务器正在运行",
  "music_dir_accessible": true,
  "music_directory": "/data/music"
}
```

### 配置验证工具

可以创建一个简单的脚本来验证环境变量：

```bash
#!/bin/bash
# verify-config.sh

echo "=== Zero Music 配置验证 ==="
echo ""

# 服务器配置
echo "📡 服务器配置:"
echo "  监听地址: ${ZERO_MUSIC_SERVER_HOST:-0.0.0.0 (默认)}"
echo "  监听端口: ${ZERO_MUSIC_SERVER_PORT:-8080 (默认)}"
echo "  读取超时: ${ZERO_MUSIC_SERVER_READ_TIMEOUT_SECONDS:-15 (默认)} 秒"
echo "  写入超时: ${ZERO_MUSIC_SERVER_WRITE_TIMEOUT_SECONDS:-60 (默认)} 秒"
echo "  空闲超时: ${ZERO_MUSIC_SERVER_IDLE_TIMEOUT_SECONDS:-120 (默认)} 秒"
echo "  关停超时: ${ZERO_MUSIC_SERVER_SHUTDOWN_TIMEOUT_SECONDS:-30 (默认)} 秒"
echo ""

# 音乐库配置
echo "🎵 音乐库配置:"
echo "  音乐目录: ${ZERO_MUSIC_MUSIC_DIRECTORY:-(未设置，将使用默认值)}"
echo "  缓存TTL: ${ZERO_MUSIC_CACHE_TTL_MINUTES:-5 (默认)} 分钟"
echo "  最大Range: ${ZERO_MUSIC_MAX_RANGE_SIZE:-104857600 (默认)} 字节"
echo ""

# 调试配置
echo "🐛 调试配置:"
echo "  调试模式: ${ZERO_MUSIC_DEBUG:-false (默认)}"
echo "  日志级别: ${LOG_LEVEL:-info (默认)}"
echo ""

# 验证音乐目录
if [ -n "$ZERO_MUSIC_MUSIC_DIRECTORY" ]; then
  if [ -d "$ZERO_MUSIC_MUSIC_DIRECTORY" ]; then
    echo "音乐目录存在: $ZERO_MUSIC_MUSIC_DIRECTORY"
  else
    echo "音乐目录不存在: $ZERO_MUSIC_MUSIC_DIRECTORY"
  fi
fi
```

使用方法：

```bash
chmod +x verify-config.sh
source production.sh  # 或 development.sh
./verify-config.sh
```

## 注意事项

### 类型验证

所有环境变量都会经过严格的类型和范围验证：

1. **端口号**: 必须是 1-65535 之间的整数
2. **超时时间**:
   - 读/写/空闲超时: 1-600 秒
   - 关停超时: 1-300 秒
3. **Range 大小**: 1 字节 - 500MB
4. **缓存 TTL**: 1-1440 分钟（24小时）
5. **音乐目录**: 必须是存在且可访问的目录

**验证失败行为**:

- 如果环境变量值格式不正确或超出范围，系统会记录警告并使用配置文件值或默认值
- 如果配置文件值也无效，应用启动将失败并显示错误信息

### 路径处理

#### 相对路径

```bash
# 相对路径会被转换为绝对路径
export ZERO_MUSIC_MUSIC_DIRECTORY=./music
# 实际使用: /opt/zero-music/music
```

#### 绝对路径

```bash
# 绝对路径直接使用
export ZERO_MUSIC_MUSIC_DIRECTORY=/data/music
# 实际使用: /data/music
```

#### 用户目录

```bash
# Linux/macOS: ~ 会被展开
export ZERO_MUSIC_MUSIC_DIRECTORY=~/Music
# 实际使用: /home/username/Music

# Windows: 不支持 ~，使用完整路径
set ZERO_MUSIC_MUSIC_DIRECTORY=C:\Users\Username\Music
```

### 安全建议

#### 1. 生产环境必须禁用调试模式

```bash
# 危险 - 可能泄露敏感信息
export ZERO_MUSIC_DEBUG=true

# 安全 - 生产环境配置
export ZERO_MUSIC_DEBUG=false
```

#### 2. 合理设置超时时间

```bash
# 过小 - 可能导致大文件传输失败
export ZERO_MUSIC_SERVER_WRITE_TIMEOUT_SECONDS=5

# 推荐 - 平衡性能和安全
export ZERO_MUSIC_SERVER_READ_TIMEOUT_SECONDS=15
export ZERO_MUSIC_SERVER_WRITE_TIMEOUT_SECONDS=60
export ZERO_MUSIC_SERVER_IDLE_TIMEOUT_SECONDS=120
```

#### 3. 限制 Range 请求大小

```bash
# 过大 - 可能导致内存溢出
export ZERO_MUSIC_MAX_RANGE_SIZE=524288000  # 500MB

# 推荐 - 适中的大小
export ZERO_MUSIC_MAX_RANGE_SIZE=104857600  # 100MB
```

#### 4. 保护敏感配置文件

```bash
# 设置适当的文件权限
chmod 600 /etc/zero-music/environment
chmod 600 .env

# 避免将 .env 文件提交到版本控制
echo ".env" >> .gitignore
```

### 性能优化建议

#### 1. 缓存 TTL 设置

根据音乐库更新频率调整：

```bash
# 高频更新（开发环境）
export ZERO_MUSIC_CACHE_TTL_MINUTES=1

# 正常更新（生产环境）
export ZERO_MUSIC_CACHE_TTL_MINUTES=5

# 低频更新（稳定环境）
export ZERO_MUSIC_CACHE_TTL_MINUTES=30
```

#### 2. 超时优化

根据网络环境和文件大小调整：

```bash
# 快速网络，小文件为主
export ZERO_MUSIC_SERVER_READ_TIMEOUT_SECONDS=10
export ZERO_MUSIC_SERVER_WRITE_TIMEOUT_SECONDS=30

# 慢速网络，大文件（FLAC/WAV）
export ZERO_MUSIC_SERVER_READ_TIMEOUT_SECONDS=30
export ZERO_MUSIC_SERVER_WRITE_TIMEOUT_SECONDS=120
```

### 日志配置

#### 日志级别说明

| 级别 | 说明 | 使用场景 |
|------|------|---------|
| `debug` | 详细的调试信息 | 开发环境，问题排查 |
| `info` | 一般信息（默认） | 生产环境 |
| `warn` | 警告信息 | 生产环境 |
| `error` | 错误信息 | 生产环境 |
| `fatal` | 致命错误 | 严重问题 |

#### 日志轮转

日志系统自动处理轮转，配置参数：

- 单文件最大: 100MB
- 保留文件数: 7 个
- 保留天数: 30 天
- 压缩: 是（gzip）

**日志文件示例**:

```text
/var/log/zero-music/
├── app.log          # 当前日志
├── app.log.1        # 前一天
├── app.log.2.gz     # 压缩的旧日志
├── app.log.3.gz
└── ...
```

### 常见问题

#### Q1: 环境变量不生效？

**可能原因**:

1. 变量名拼写错误
2. 值格式不正确（如端口号包含非数字字符）
3. 值超出有效范围
4. 环境变量未正确导出（缺少 `export`）

**解决方法**:

```bash
# 检查环境变量是否已设置
echo $ZERO_MUSIC_SERVER_PORT

# 确保使用 export
export ZERO_MUSIC_SERVER_PORT=8080  
ZERO_MUSIC_SERVER_PORT=8080         # 不会传递给子进程

# 查看所有 ZERO_MUSIC 相关的环境变量
env | grep ZERO_MUSIC
```

#### Q2: 音乐目录配置后找不到文件？

**检查步骤**:

1. 验证目录存在且可访问
2. 检查目录权限
3. 确认路径是绝对路径

```bash
# 验证目录
ls -la $ZERO_MUSIC_MUSIC_DIRECTORY

# 检查权限
# 确保运行用户有读取权限
sudo -u zeromusic ls $ZERO_MUSIC_MUSIC_DIRECTORY
```

#### Q3: Docker 容器中环境变量不生效？

**Docker 特殊处理**:

```bash
# 确保环境变量正确传递
docker run -e ZERO_MUSIC_DEBUG=false ...  # 

# 或使用 --env-file
docker run --env-file .env ...  # 

# 检查容器内环境变量
docker exec zero-music env | grep ZERO_MUSIC
```

### 推荐配置方案

#### 小型部署（个人使用）

```bash
export ZERO_MUSIC_SERVER_PORT=8080
export ZERO_MUSIC_MUSIC_DIRECTORY=~/Music
export ZERO_MUSIC_CACHE_TTL_MINUTES=10
export ZERO_MUSIC_DEBUG=false
export LOG_LEVEL=info
```

#### 中型部署（团队使用）

```bash
export ZERO_MUSIC_SERVER_HOST=0.0.0.0
export ZERO_MUSIC_SERVER_PORT=8080
export ZERO_MUSIC_SERVER_READ_TIMEOUT_SECONDS=15
export ZERO_MUSIC_SERVER_WRITE_TIMEOUT_SECONDS=60
export ZERO_MUSIC_SERVER_IDLE_TIMEOUT_SECONDS=120
export ZERO_MUSIC_MUSIC_DIRECTORY=/data/music
export ZERO_MUSIC_MAX_RANGE_SIZE=104857600
export ZERO_MUSIC_CACHE_TTL_MINUTES=5
export ZERO_MUSIC_DEBUG=false
export LOG_LEVEL=info
```

#### 大型部署（企业使用）

```bash
export ZERO_MUSIC_SERVER_HOST=0.0.0.0
export ZERO_MUSIC_SERVER_PORT=8080
export ZERO_MUSIC_SERVER_READ_TIMEOUT_SECONDS=20
export ZERO_MUSIC_SERVER_WRITE_TIMEOUT_SECONDS=90
export ZERO_MUSIC_SERVER_IDLE_TIMEOUT_SECONDS=180
export ZERO_MUSIC_SERVER_SHUTDOWN_TIMEOUT_SECONDS=60
export ZERO_MUSIC_MUSIC_DIRECTORY=/mnt/nfs/music
export ZERO_MUSIC_MAX_RANGE_SIZE=52428800  # 50MB 更保守
export ZERO_MUSIC_CACHE_TTL_MINUTES=3      # 更频繁刷新
export ZERO_MUSIC_DEBUG=false
export LOG_LEVEL=warn  # 减少日志量
```

## 相关文档

- [配置文件说明](../config.json) - JSON 配置文件示例
- [API 文档](./API.md) - REST API 接口文档
- [部署指南](./DEPLOYMENT.md) - 生产环境部署指南
- [故障排查](./TROUBLESHOOTING.md) - 常见问题解决方案

## 更新历史

- **2025-11-23**:
  - 新增 HTTP 超时配置（READ/WRITE/IDLE/SHUTDOWN）
  - 新增调试模式配置（ZERO_MUSIC_DEBUG）
  - 新增日志级别配置（LOG_LEVEL）
  - 完善配置示例和最佳实践
  - 添加 Docker、systemd 等部署方式的配置说明
- **初始版本**: 基础环境变量支持
