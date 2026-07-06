#!/bin/bash
# 完整清空 investment-kb 所有输出
# 用法: ./cleanup.sh [vault_path] [config_path]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CONFIG_PATH="${2:-$PROJECT_ROOT/config.yaml}"

# 从 config.yaml 读取 vault_path
if [ -z "$1" ]; then
    VAULT_PATH=$(grep "^obsidian_vault_path:" "$CONFIG_PATH" 2>/dev/null | head -1 | sed -E "s/^obsidian_vault_path:[[:space:]]*['\"]?([^'\"]*)['\"]?/\1/" | sed 's/\\/\//g')
else
    VAULT_PATH="$1"
fi

# 校验 VAULT_PATH 不能为空
if [ -z "$VAULT_PATH" ]; then
    echo "❌ 无法从 config.yaml 读取 obsidian_vault_path，请检查：$CONFIG_PATH"
    exit 1
fi

if [ ! -d "$VAULT_PATH" ]; then
    echo "❌ Vault 路径不存在：$VAULT_PATH"
    exit 1
fi

echo "=== 开始清空 investment-kb 输出 ==="
echo "Vault path: $VAULT_PATH"
echo "Config path: $CONFIG_PATH"
echo ""

# 1. 清空 RAW 单文件
echo "1. 清空 RAW 单文件..."
RAW_DIR="$VAULT_PATH/日常随笔/股市学习/宽基指数仓位管理系统/01-源文档/问答"
if [ -d "$RAW_DIR" ]; then
    find "$RAW_DIR" -name "RAW-*.md" -delete 2>/dev/null || true
    rm -f "$RAW_DIR/原始材料索引.md" 2>/dev/null || true
    echo "   ✓ 已清空 RAW 单文件和索引"
fi

# 2. 清空 RAW 聚合库
echo "2. 清空 RAW 聚合库..."
rm -f "$VAULT_PATH/日常随笔/股市学习/宽基指数仓位管理系统/01-源文档/问答/原始材料库.md" 2>/dev/null || true
echo "   ✓ 已清空 RAW 聚合库"

# 3. 清空 QA 单文件
echo "3. 清空 QA 单文件..."
QA_DIR="$VAULT_PATH/日常随笔/股市学习/宽基指数仓位管理系统/02-观点/问答知识卡片"
if [ -d "$QA_DIR" ]; then
    find "$QA_DIR" -name "QA-*.md" -delete 2>/dev/null || true
    rm -f "$QA_DIR/问答知识卡片索引.md" 2>/dev/null || true
    echo "   ✓ 已清空 QA 单文件和索引"
fi

# 4. 清空 QA 聚合库
echo "4. 清空 QA 聚合库..."
rm -f "$VAULT_PATH/日常随笔/股市学习/宽基指数仓位管理系统/02-观点/问答知识卡片库.md" 2>/dev/null || true
echo "   ✓ 已清空 QA 聚合库"

# 5. 清空 KNOW 单文件
echo "5. 清空 KNOW 单文件..."
KNOW_DIR="$VAULT_PATH/日常随笔/股市学习/宽基指数仓位管理系统/02-观点/宏观理解卡"
if [ -d "$KNOW_DIR" ]; then
    find "$KNOW_DIR" -name "KNOW-*.md" -delete 2>/dev/null || true
    rm -f "$KNOW_DIR/宏观理解卡索引.md" 2>/dev/null || true
    echo "   ✓ 已清空 KNOW 单文件和索引"
fi

# 6. 清空 MO 单文件
echo "6. 清空 MO 单文件..."
MO_DIR="$VAULT_PATH/日常随笔/股市学习/宽基指数仓位管理系统/02-观点/市场观察卡"
if [ -d "$MO_DIR" ]; then
    find "$MO_DIR" -name "MO-*.md" -delete 2>/dev/null || true
    rm -f "$MO_DIR/市场观察卡索引.md" 2>/dev/null || true
    echo "   ✓ 已清空 MO 单文件和索引"
fi

# 7. 清空 CR 单文件
echo "7. 清空 CR 单文件..."
CR_DIR="$VAULT_PATH/日常随笔/股市学习/宽基指数仓位管理系统/03-规则/候选规则"
if [ -d "$CR_DIR" ]; then
    find "$CR_DIR" -name "CR-*.md" -delete 2>/dev/null || true
    rm -f "$CR_DIR/候选规则索引.md" 2>/dev/null || true
    rm -f "$CR_DIR/*.backup" 2>/dev/null || true
    echo "   ✓ 已清空 CR 单文件和索引"
fi

# 8. 清空 CR 聚合库
echo "8. 清空 CR 聚合库..."
rm -f "$VAULT_PATH/日常随笔/股市学习/宽基指数仓位管理系统/03-规则/候选规则库.md" 2>/dev/null || true
echo "   ✓ 已清空 CR 聚合库"

# 9. 清空验证卡
echo "9. 清空验证卡..."
VC_DIR="$VAULT_PATH/日常随笔/股市学习/宽基指数仓位管理系统/03-规则/规则回溯验证/规则验证卡"
if [ -d "$VC_DIR" ]; then
    find "$VC_DIR" -name "*.md" -delete 2>/dev/null || true
    echo "   ✓ 已清空验证卡"
fi

# 10. 清空案例库
echo "10. 清空案例库..."
rm -f "$VAULT_PATH/日常随笔/股市学习/宽基指数仓位管理系统/03-规则/规则回溯验证/历史案例库/AI提取案例素材库.md" 2>/dev/null || true
echo "   ✓ 已清空案例库"

# 11. 清空数据处理文件
echo "11. 清空数据处理文件..."
rm -f "$PROJECT_ROOT/data/id_state.json" 2>/dev/null || true
rm -f "$PROJECT_ROOT/data/import_hashes.json" 2>/dev/null || true
rm -rf "$PROJECT_ROOT/data/output/"* 2>/dev/null || true
rm -rf "$PROJECT_ROOT/data/error_outputs/"* 2>/dev/null || true
rm -rf "$PROJECT_ROOT/data/debug/"* 2>/dev/null || true
echo "   ✓ 已清空 data 目录"

echo ""
echo "=== 清空完成 ==="
echo "现在可以重新运行 extract 命令"
