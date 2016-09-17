# genext2fs-tar-index

`genext2fs-tar-index` is a pretty specialized tool that takes a `tar`
archive of the desired contents of an `ext2` filesystem and produces
a file information table in the format expected by
[`genext2fs`](https://github.com/devcurmudgeon/genext2fs).

`genext2fs` is a program for creating ext2 filesystem images without
root access and without mounting anything. This can be useful as
part of an automated build process for a Linux distribution, for
example.

`genext2fs` reads from a directory containing the files desired in the
target filesystem, but non-root users are not able to create device nodes
and files owned by other userids, so the file information table is used
to provide this extra metadata that the local filesystem cannot represent.

This was written with the goal of using it as part of a pipeline for
making disk images containing embedded linux installations produced using
[Buildroot](https://buildroot.org/). One of Buildroot's options is to produce
a `rootfs.tar` containing the root filesystem to be installed on the target.
Given this, we can produce a real ext2 filesystem of those files as follows:

* `mkdir rootfs`
* `tar xvf rootfs.tar -C rootfs --no-acls --no-same-owner --no-same-permissions --no-selinux --no-xattrs` (extract the root filesystem image, discarding ownership/permissions)
* `genext2fs-tar-index rootfs.tar >rootfs.tab` (produce table of the ownership and permissions of contained files)
* `genext2fs -b 1M -d rootfs -D rootfs.tab root.img` (produce ~1GB ext2 filesystem image with the desired contents)

After this some other steps are optional but suggested:

* `tune2fs -j root.img` (enable journal, effectively turning this into ext3)
* `e2fsck -yfD root.img` (fix up inconsistencies added by previous step, making it *really* ext3, and optimize)

From here you might use something like
[`genimage`](https://github.com/vivien/genimage) to produce a whole-disk image
with a partition table, and write to it some sort of bootloader to produce
a bootable system.
