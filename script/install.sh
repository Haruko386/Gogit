install_dir="$(go env GOBIN)"
if [ -z "$install_dir" ]; then
    install_dir="$(go env GOPATH)/bin"
fi

mkdir -p "$install_dir"
go build -o "$install_dir/gogit" .