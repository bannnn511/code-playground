arch=$(uname -m)
dest_kernel="hello-vmlinux.bin"
dest_rootfs="hello-rootfs.ext4"
image_bucket_url="https://s3.amazonaws.com/spec.ccfc.min/img/quickstart_guide/$arch"

if [ ${arch} = "x86_64" ]; then
    kernel="${image_bucket_url}/kernels/vmlinux.bin"
    rootfs="${image_bucket_url}/rootfs/bionic.rootfs.ext4"
elif [ ${arch} = "aarch64" ]; then
    kernel="${image_bucket_url}/kernels/vmlinux.bin"
    rootfs="${image_bucket_url}/rootfs/bionic.rootfs.ext4"
else
    echo "Cannot run firecracker on $arch architecture!"
    exit 1
fi

if [ ! -f $dest_kernel ]; then
    echo "Kernel not found, downloading $kernel..."
    curl -fsSL -o $dest_kernel $kernel
    echo "Saved kernel file to $dest_kernel."
fi

if [ ! -f $dest_rootfs ]; then
    echo "Rootfs not found, downloading $rootfs..."
    curl -fsSL -o $dest_rootfs $rootfs
    echo "Saved root block device to $dest_rootfs."
fi

echo "Downloading public key file..."
[ -e hello-id_rsa ] || wget -O hello-id_rsa https://raw.githubusercontent.com/firecracker-microvm/firecracker-demo/ec271b1e5ffc55bd0bf0632d5260e96ed54b5c0c/xenial.rootfs.id_rsa
echo "Saved public key file."
