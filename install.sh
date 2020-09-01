# This script is adapted from the deno project's deno_install repo:
# https://github.com/qri-io/qri_install/blob/master/install.sh
# Deno is a great project, you should check it out.

set -e

case $(uname -s) in
Darwin) os="darwin" ;;
*) os="linux" ;;
esac

case $(uname -m) in
x86_64) arch="amd64" ;;
*) arch="other" ;;
esac

if [ "$arch" = "other" ]; then
  # TODO (b5) - support other archs. We do.
	echo "Unsupported architecture for install script: $(uname -m). Please download manually from github.com/qri-io/qri/releases"
	exit
fi

if [ $# -eq 0 ]; then
	qri_asset_path=$(
		command curl -sSf https://github.com/qri-io/qri/releases |
			command grep -o "/qri-io/qri/releases/download/.*/qri_${os}_amd64\\.zip" |
			command head -n 1
	)
	if [ ! "$qri_asset_path" ]; then exit 1; fi
	qri_uri="https://github.com${qri_asset_path}"
else
	qri_uri="https://github.com/qri-io/qri/releases/download/${1}/qri_${os}_amd64.zip"
fi

qri_install="${QRI_INSTALL:-/usr/local}"
bin_dir="$qri_install/bin"
exe="$bin_dir/qri"

if [ ! -d "$bin_dir" ]; then
	mkdir -p "$bin_dir"
fi

curl --fail --location --progress-bar --output "$exe.zip" "$qri_uri"
unzip -d "$bin_dir" "$exe.zip"
chmod +x "$exe"
rm "$bin_dir/readme.md" "$exe.zip"

echo "Qri was installed successfully to $exe"
if command -v qri >/dev/null; then
	echo "Run 'qri --help' to get started"
else
	echo "Manually add the directory to your \$HOME/.bash_profile (or similar)"
	echo "  export QRI_INSTALL=\"$qri_install\""
	echo "  export PATH=\"\$QRI_INSTALL/bin:\$PATH\""
	echo "Run '$exe --help' to get started"
fi
