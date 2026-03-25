import pytest

from novabackup.core import normalize_vm_type, list_vms


def test_normalize_vm_type_various_inputs():
    assert normalize_vm_type("Hyper-V") == "HyperV"
    assert normalize_vm_type("hyper-v") == "HyperV"
    assert normalize_vm_type("KVM") == "KVM"
    assert normalize_vm_type("kvm") == "KVM"
    assert normalize_vm_type("vmware") == "VMware"
    assert normalize_vm_type("VMware-workstation") == "VMware"
    assert normalize_vm_type("unknown") == "Unknown"


def test_list_vms_returns_list_of_dicts():
    vms = list_vms()
    assert isinstance(vms, list)
    assert len(vms) >= 1
    for vm in vms:
        assert isinstance(vm, dict)
        assert "id" in vm and "name" in vm and "type" in vm and "status" in vm
