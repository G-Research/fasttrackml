#!/bin/sh

set -e

case $(uname -s) in
    "Darwin")
        reverse="tail -r"
        os="macos"
        ;;
    "Linux")
        reverse="tac"
        os="linux"
        ;;
    *)
        echo "Unsupported OS"
        exit 1
        ;;
esac

case $(uname -m) in
    "x86_64")
        arch="x86_64"
        ;;
    "arm64" | "aarch64")
        arch="arm64"
        ;;
    *)
        echo "Unsupported architecture"
        exit 1
        ;;
esac

fetch() {
    if which curl >/dev/null; then
        if [ "$#" -eq 2 ]; then curl -L -o "$1" "$2"; else curl -sSL "$1"; fi
    elif which wget >/dev/null; then
        if [ "$#" -eq 2 ]; then wget -O "$1" "$2"; else wget -nv -O - "$1"; fi
    else
        echo "Can't find curl or wget, can't download package"
        exit 1
    fi
}

target="${os}_${arch}"
echo "Detected target: $target"

url=$(
    # We reverse the output twice to buffer the entire output of curl/wget and
    # avoid errors when the pipe is closed early.
    fetch https://api.github.com/repos/G-Research/fasttrackml/releases/latest |
        $reverse | $reverse | grep -wo -m1 "https://.*$target.tar.gz" || true
)
if ! test "$url"; then
    echo "Could not find release info"
    exit 1
fi

echo "Downloading FastTrackML..."

temp_dir=$(mktemp -dt fml.XXXXXX)
trap 'rm -rf "$temp_dir"' EXIT INT TERM
cd "$temp_dir"

if ! fetch fml.tar.gz "$url"; then
    echo "Could not download tarball"
    exit 1
fi

user_bin="$HOME/.local/bin"
case $PATH in
    *:"$user_bin":* | "$user_bin":* | *:"$user_bin")
        default_bin=$user_bin
        ;;
    *)
        default_bin='/usr/local/bin'
        ;;
esac

_read_installdir() {
    printf "Install location [default: %s]: " "$default_bin"
    read -r fml_installdir </dev/tty
    fml_installdir=${fml_installdir:-$default_bin}
}

if [ -z "$FML_BINDIR" ]; then
    _read_installdir

    while ! test -d "$fml_installdir"; do
        echo "Directory $fml_installdir does not exist"
        _read_installdir
    done
else
    fml_installdir=${FML_BINDIR}
fi

tar xzf fml.tar.gz

if test -w "$fml_installdir" || [ -n "$FML_BINDIR" ]; then
    mv fml "$fml_installdir/"
else
    sudo mv fml "$fml_installdir/"
fi

echo "$("$fml_installdir"/fml --version) has been installed to $fml_installdir/fml"
