#!/bin/bash

# SecurityMCP-Hub 构建脚本
# 用于构建所有 MCP 服务

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SERVERS_DIR="$PROJECT_ROOT/servers"

echo "开始构建 SecurityMCP-Hub 服务..."

# 遍历所有服务目录
for service_dir in "$SERVERS_DIR"/*; do
    if [ -d "$service_dir" ] && [ -f "$service_dir/server.go" ]; then
        service_name=$(basename "$service_dir")
        echo ""
        echo "构建服务: $service_name"
        echo "----------------------------------------"
        
        cd "$service_dir"
        
        if [ -f "go.mod" ]; then
            echo "编译 $service_name..."
            go build -o "$service_name" server.go
            if [ $? -eq 0 ]; then
                echo "✓ $service_name 构建成功"
            else
                echo "✗ $service_name 构建失败"
            fi
        else
            echo "跳过 $service_name (不是 Go 项目)"
        fi
        
        cd "$PROJECT_ROOT"
    fi
done

echo ""
echo "构建完成！"

