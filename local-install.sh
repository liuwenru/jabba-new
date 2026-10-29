#!/usr/bin/env bash
# 本地安装 jabba：不下载、不编译，只配置 shell 集成。
#
# 前提: 已把 jabba 二进制放到 $JABBA_HOME/bin/jabba（或 $JABBA_HOME/jabba）
# 用法: ./local-install.sh
#   JABBA_HOME 默认为本脚本所在目录
set -euo pipefail

JABBA_HOME=${JABBA_HOME:-$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)}
JABBA_BIN="$JABBA_HOME/bin/jabba"

# 兼容二进制直接放在 JABBA_HOME 根目录的情况（视为新版本，覆盖 bin 下的旧版本）
if [ -f "$JABBA_HOME/jabba" ]; then
    mkdir -p "$JABBA_HOME/bin"
    mv -f "$JABBA_HOME/jabba" "$JABBA_BIN"
fi
[ -f "$JABBA_BIN" ] || { echo "[ERROR] 未找到 $JABBA_BIN" >&2; exit 1; }
chmod +x "$JABBA_BIN"
"$JABBA_BIN" --version >/dev/null 2>&1 || { echo "[ERROR] $JABBA_BIN 无法运行，请检查平台/架构" >&2; exit 1; }

# jabba 通过 fd3 输出要在当前 shell 执行的命令（export JAVA_HOME/PATH 等），需要函数包一层 eval
cat >"$JABBA_HOME/jabba.sh" <<EOF
export JABBA_HOME="$JABBA_HOME"

jabba() {
    local fd3=\$(mktemp /tmp/jabba-fd3.XXXXXX)
    (JABBA_SHELL_INTEGRATION=ON "\$JABBA_HOME/bin/jabba" "\$@" 3>| \${fd3})
    local exit_code=\$?
    eval \$(cat \${fd3})
    rm -f \${fd3}
    return \${exit_code}
}

if [ -n "\$(jabba alias default)" ]; then
    jabba use default
fi
EOF
cat >>"$JABBA_HOME/jabba.sh" <<'EOF'

# jabba shell completion (managed by jabba init)
if [ -n "${BASH_VERSION:-}" ]; then
    case $- in
        *i*)
            if ! declare -F _get_comp_words_by_ref >/dev/null; then
                for _jabba_completion_file in /usr/share/bash-completion/bash_completion "${HOMEBREW_PREFIX:-/opt/homebrew}/etc/profile.d/bash_completion.sh" /usr/local/etc/profile.d/bash_completion.sh /usr/local/etc/bash_completion; do
                    if [ -r "$_jabba_completion_file" ]; then
                        . "$_jabba_completion_file"
                        break
                    fi
                done
                unset _jabba_completion_file
            fi
            if declare -F _get_comp_words_by_ref >/dev/null; then
                eval "$(jabba completion bash)"
            fi
            ;;
    esac
elif [ -n "${ZSH_VERSION:-}" ]; then
    if [[ -o interactive ]]; then
        if ! (( $+functions[compdef] )); then
            autoload -Uz compinit
            compinit
        fi
        if (( $+functions[compdef] )); then
            eval "$(jabba completion zsh)"
        fi
    fi
fi
EOF

SOURCE_LINE="[ -s \"$JABBA_HOME/jabba.sh\" ] && source \"$JABBA_HOME/jabba.sh\""
for rc in "$HOME/.bashrc" "$HOME/.zshrc"; do
    [ "$rc" = "$HOME/.zshrc" ] && ! command -v zsh >/dev/null 2>&1 && continue
    touch "$rc"
    if grep -q '/jabba.sh' "$rc"; then
        echo "跳过 $rc（已配置）"
    else
        printf '\n%s\n' "$SOURCE_LINE" >>"$rc"
        echo "已写入 $rc"
    fi
done

echo "安装完成: $("$JABBA_BIN" --version)。重新打开终端，或执行: source \"$JABBA_HOME/jabba.sh\""
