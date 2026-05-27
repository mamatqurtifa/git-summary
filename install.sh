#!/usr/bin/env bash
# install.sh — build and install git-summary + shell completion

set -e

BINARY="git-summary"
INSTALL_DIR="${PREFIX:-/usr/local/bin}"

echo "Building $BINARY..."
go build -ldflags="-s -w" -o "$BINARY" .

echo "Installing to $INSTALL_DIR/$BINARY ..."
if [ -w "$INSTALL_DIR" ]; then
    mv "$BINARY" "$INSTALL_DIR/"
else
    sudo mv "$BINARY" "$INSTALL_DIR/"
fi

echo "Setting up shell completion..."

SHELL_NAME=$(basename "$SHELL")

case "$SHELL_NAME" in
    bash)
        COMP_FILE="$HOME/.bash_completion.d/git-summary"
        mkdir -p "$HOME/.bash_completion.d"
        "$INSTALL_DIR/$BINARY" completion bash > "$COMP_FILE"
        # Add sourcing to .bashrc if not already there
        BASHRC="$HOME/.bashrc"
        SOURCE_LINE="source $COMP_FILE"
        if ! grep -qF "$COMP_FILE" "$BASHRC" 2>/dev/null; then
            echo "" >> "$BASHRC"
            echo "# git-summary completion" >> "$BASHRC"
            echo "$SOURCE_LINE" >> "$BASHRC"
            echo "  Added completion to $BASHRC"
        fi
        echo "  Bash completion installed. Run: source ~/.bashrc"
        ;;
    zsh)
        ZSH_COMP_DIR="${ZDOTDIR:-$HOME}/.zsh/completions"
        mkdir -p "$ZSH_COMP_DIR"
        "$INSTALL_DIR/$BINARY" completion zsh > "$ZSH_COMP_DIR/_git-summary"
        ZSHRC="${ZDOTDIR:-$HOME}/.zshrc"
        FPATH_LINE="fpath=($ZSH_COMP_DIR \$fpath)"
        if ! grep -qF "$ZSH_COMP_DIR" "$ZSHRC" 2>/dev/null; then
            echo "" >> "$ZSHRC"
            echo "# git-summary completion" >> "$ZSHRC"
            echo "$FPATH_LINE" >> "$ZSHRC"
            echo "autoload -U compinit && compinit" >> "$ZSHRC"
        fi
        echo "  Zsh completion installed. Run: source ~/.zshrc"
        ;;
    fish)
        FISH_COMP="$HOME/.config/fish/completions/git-summary.fish"
        mkdir -p "$(dirname $FISH_COMP)"
        "$INSTALL_DIR/$BINARY" completion fish > "$FISH_COMP"
        echo "  Fish completion installed automatically."
        ;;
    *)
        echo "  Unknown shell '$SHELL_NAME'. Manually install completion:"
        echo "    git-summary completion bash|zsh|fish"
        ;;
esac

echo ""
echo "Done! Try: git-summary --help"
