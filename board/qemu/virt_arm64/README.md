TamaGo - bare metal Go - QEMU virt ARM64 VM support
====================================================

The `virt_arm64` package provides an ARM64 QEMU `virt` board target for local
TamaGo VM testing. It is intentionally separate from the Cloud Hypervisor ARM64
board because the GIC and PCI maps differ.

Current scope
=============

This target is intended for direct-kernel QEMU smoke tests:

* RAM starts at `0x40000000`.
* GICv3 distributor is at `0x08000000`.
* GICv3 redistributor is at `0x080a0000`.
* GICv3 ITS is at `0x08080000`.
* PL011 UART is at `0x09000000`.
* The first virtio-mmio transport is at `0x0a000000`.

The constants match the FDT emitted by:

```sh
qemu-system-aarch64 -machine virt,gic-version=3,dumpdtb=virt.dtb
```

Applications should import this board package:

```go
import _ "github.com/usbarmory/tamago/board/qemu/virt_arm64"
```
