TamaGo ARM64 VM Upstreaming Plan
================================

This branch carries experimental support for small `GOOS=tamago GOARCH=arm64`
appliances under QEMU `virt` and Cloud Hypervisor. The work should be proposed
upstream as small, reviewable patch groups rather than as one branch.

Patch groups
============

1. QEMU `virt` ARM64 board
--------------------------

Scope:

* `board/qemu/virt_arm64`
* PL011 console mapping.
* QEMU `virt` GICv3 memory map.
* First virtio-mmio transport discovery.
* Board README with import/build guidance.

Validation:

* `armory-ssh-agent` QEMU `virt` ARM64 smoke boots locally.
* Host-forwarded HTTP fingerprint check passes through virtio-mmio networking.

2. Cloud Hypervisor ARM64 board
-------------------------------

Scope:

* `board/cloud_hypervisor/arm64`
* Cloud Hypervisor RAM, PL011, GICv3, and ECAM constants.
* 128 MB non-heap DMA carveout for virtio queues.
* Board README with AArch64 Linux `Image` boot guidance.

Validation:

* Cloud Hypervisor accepts the AArch64 `Image` wrapper.
* A Fedora Asahi ARM64 KVM host reaches post-MMU TamaGo execution.
* Tap-backed virtio-net HTTP fingerprint smoke passes.

3. VirtIO and PCI compatibility fixes
-------------------------------------

Scope:

* Preserve virtio transport/reserved feature bits during feature negotiation.
* Mask PCI BAR low flag bits before returning MMIO base addresses.
* Read virtio PCI config space byte-by-byte to avoid ARM device-memory
  `memmove` faults.

Validation:

* Required for Cloud Hypervisor ARM64 virtio-net initialization and packet I/O.
* Does not change board-level APIs.

4. Cloud Hypervisor ARM64 virtio-net BAR assignment
---------------------------------------------------

Scope:

* Assign the Cloud Hypervisor ARM64 virtio-net 64-bit BAR into guest MMIO space.
* Enable PCI memory-space and bus-master bits.

Validation:

* Required for tap-backed Cloud Hypervisor ARM64 network smoke.

Known follow-ups
================

* Validate ARM64 Cloud Hypervisor on more host kernels and Cloud Hypervisor
  versions.
* Validate SpectrumOS nested VM execution separately from the local Asahi host.
* Validate any amd64 regressions for the virtio/PCI compatibility fixes.
* Decide whether transitional virtio-net compatibility should be included in the
  initial upstream series or left as a follow-up.
* Add explicit cache maintenance if future DMA mappings become cacheable.

Current downstream appliance validation
=======================================

The `armory-ssh-agent` downstream appliance branch validates:

* QEMU `virt` ARM64 boot and HTTP fingerprint smoke.
* Cloud Hypervisor ARM64 no-net boot smoke.
* Cloud Hypervisor ARM64 tap-backed network smoke.
* USB Armory Mk II hardware CDC ECM, HTTP, eMMC persistence status, and
  SSH-agent signing smoke.
