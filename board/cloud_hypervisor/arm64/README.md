TamaGo - bare metal Go - Cloud Hypervisor ARM64 VM support
==========================================================

The `arm64` package provides initial support for Cloud Hypervisor style ARM64
virtual machines and QEMU `virt` direct-kernel bring-up.

Current scope
=============

This target is an ARM64 VM bring-up target for small TamaGo appliances:

* RAM starts at `0x40000000`.
* GICv3 distributor is at `0x08000000`.
* GICv3 redistributor is at `0x080a0000`.
* PL011 UART is at `0x09000000`.
* PCI ECAM starts at `0x4010000000`.
* Modern VirtIO PCI devices can be discovered through ECAM.

These constants match the ARM64 `virt` environment observed in local
Spectrum/QEMU validation logs. They are expected to be compatible with the
Cloud Hypervisor ARM64 direct-kernel environment, but runtime validation is
still required.

Validation status
=================

The current implementation has been validated under QEMU ARM64 direct-kernel
boot with a modern `virtio-net-pci` device and user-mode host forwarding. A
TamaGo HTTP/SSH-agent appliance can answer HTTP requests through the virtio-net
path.

The board reserves the top 128MB of guest RAM as a non-heap DMA carveout and
maps it as Normal non-cacheable memory. This avoids overlapping DMA buffers with
the Go heap while still allowing byte access to virtqueue memory.

The validated QEMU shape is:

```sh
qemu-system-aarch64 \
  -machine virt,gic-version=3 -cpu max -m 512M \
  -nographic -monitor none -serial stdio \
  -netdev user,id=n0,net=10.0.0.0/24,host=10.0.0.2,hostfwd=tcp::10080-10.0.0.1:80 \
  -device virtio-net-pci,netdev=n0,disable-legacy=on,disable-modern=off \
  -kernel app.elf
```

Remaining work
==============

The next upstreamable increment is ARM64 MSI/MSI-X or legacy INTx routing
suitable for virtio-pci devices, plus explicit cache maintenance if cacheable
DMA mappings are introduced later.

Additional validation is still needed on Cloud Hypervisor itself, SpectrumOS
nested virtual machines, AMD64 hosts, and ARM64 hosts. Spectrum/QEMU logs also
show transitional virtio device IDs in some configurations, so compatibility
with transitional virtio-net remains a separate follow-up.

Compiling
=========

Applications should import this board package:

```go
import _ "github.com/usbarmory/tamago/board/cloud_hypervisor/arm64"
```

Then build with a TamaGo-enabled Go toolchain:

```sh
GOOS=tamago GOOSPKG=github.com/usbarmory/tamago GOARCH=arm64 \
	go build -ldflags "-T 0x40080000 -R 0x1000"
```
