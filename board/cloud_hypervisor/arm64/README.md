TamaGo - bare metal Go - Cloud Hypervisor ARM64 VM support
==========================================================

The `arm64` package provides initial support for Cloud Hypervisor style ARM64
virtual machines.

Current scope
=============

This target is an ARM64 VM bring-up target for small TamaGo appliances:

* RAM starts at `0x40000000`.
* GICv3 distributor is at `0x08ff0000`.
* GICv3 redistributor is at `0x08fd0000`.
* GICv3 ITS is at `0x08fb0000`.
* PL011 UART is at `0x09000000`.
* PCI ECAM starts at `0x30000000`.
* Modern VirtIO PCI devices can be discovered through ECAM.

These constants match the ARM64 FDT emitted by Cloud Hypervisor.

Validation status
=================

The current implementation has been validated far enough for Cloud Hypervisor
to accept an AArch64 Linux `Image` wrapper and start the vCPU with the expected
FDT memory map.

On a Fedora Asahi ARM64 host with 16K pages, execution currently reaches
TamaGo's ARM64 `Hwinit0` under Cloud Hypervisor/KVM and stalls while enabling
`SCTLR_EL1.M` in `InitMMU`. Temporary PL011 breadcrumbs showed the VM reaches
the EL2-to-EL1 drop, `Hwinit0`, `fp_enable`, `TTBR0_EL1`, and TLB invalidation
before the first post-MMU instruction fails to complete.

The same Image reaches the post-MMU path under QEMU `virt` with TCG, confirming
that the Linux Image wrapper and the existing 4K-granule TamaGo translation
tables are structurally usable in an ARM64 VM. QEMU `virt` then faults later in
`GIC.Init()` because this Cloud Hypervisor board intentionally uses the Cloud
Hypervisor GIC register map, not QEMU's `virt` map.

An experimental 16K-granule table layout and a 48-bit `TCR_EL1.IPS` setting did
not move the Cloud Hypervisor/KVM path past MMU enable on that host, so the
current blocker is more specific than simply converting TamaGo to 16K
translation tables.

End-to-end Cloud Hypervisor packet delivery still needs to be validated after
the ARM64 KVM MMU-enable issue is resolved.

The board reserves the top 128MB of guest RAM as a non-heap DMA carveout and
maps it as Normal non-cacheable memory. This avoids overlapping DMA buffers with
the Go heap while still allowing byte access to virtqueue memory.

For direct-kernel boot, build a Linux-style AArch64 `Image` wrapper around the
TamaGo ELF. Cloud Hypervisor expects the ARM64 Image header rather than a raw
ELF.

The minimal Cloud Hypervisor smoke shape is:

```sh
cloud-hypervisor \
  --memory size=512M \
  --kernel app.Image \
  --console off \
  --serial tty
```

Remaining work
==============

The next upstreamable increment is ARM64 MSI/MSI-X or legacy INTx routing
suitable for virtio-pci devices, plus explicit cache maintenance if cacheable
DMA mappings are introduced later.

Additional validation is still needed for Cloud Hypervisor virtio-net,
SpectrumOS nested virtual machines, AMD64 hosts, and ARM64 hosts. QEMU `virt`
uses a different GIC/ECAM map and should be kept as a separate board target.
Spectrum/QEMU logs also show transitional virtio device IDs in some
configurations, so compatibility with transitional virtio-net remains a
separate follow-up.

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
