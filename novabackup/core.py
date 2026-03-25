from typing import List, Dict

import platform


def normalize_vm_type(vm_type: str) -> str:
    """Normalize a VM type to a canonical form."""
    if not vm_type:
        return vm_type
    t = vm_type.strip().lower()
    if t in {"hyperv", "hyper-v"}:
        return "HyperV"
    if t in {"kvm", "qemu"}:
        return "KVM"
    if t in {"vmware", "vmware-workstation"}:
        return "VMware"
    return vm_type.title()


# Try import a concrete provider (Hyper-V). If not available, we'll fallback to mock data.
try:
    from .providers.hyperv import HyperVProvider  # type: ignore
except Exception:
    HyperVProvider = None  # type: ignore

# Try import a concrete Linux provider (Libvirt). If not available, we'll fallback to mock data.
try:
    from .providers.libvirt import LibvirtProvider  # type: ignore
except Exception:
    LibvirtProvider = None  # type: ignore


def list_vms() -> List[Dict[str, str]]:
    """Return a list of VMs from a real provider or a mock list if not available."""
    vms: List[Dict[str, str]] = []
    provider = None
    if HyperVProvider is not None:
        try:
            provider = HyperVProvider()
            vms = provider.list_vms()
        except Exception:
            vms = []

    # If Hyper-V provider failed or is not available, try Libvirt provider (Linux)
    if (not isinstance(vms, list)) or (len(vms) == 0):
        if LibvirtProvider is not None:
            try:
                provider = LibvirtProvider()
                vms = provider.list_vms()
            except Exception:
                vms = []

    if not isinstance(vms, list) or len(vms) == 0:
        vms = [
            {
                "id": "vm1",
                "name": "win10-dev",
                "type": normalize_vm_type("Hyper-V"),
                "status": "running",
            },
            {
                "id": "vm2",
                "name": "ubuntu22",
                "type": normalize_vm_type("Hyper-V"),
                "status": "stopped",
            },
        ]
    return vms
