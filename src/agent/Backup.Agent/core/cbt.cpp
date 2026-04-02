#include "cbt.h"

namespace backup {

HyperVCBTProvider::HyperVCBTProvider() = default;

bool HyperVCBTProvider::EnableCBT(const std::string& vm_id) {
    enabled_ = true;
    return true;
}

bool HyperVCBTProvider::DisableCBT(const std::string& vm_id) {
    enabled_ = false;
    return true;
}

bool HyperVCBTProvider::IsCBTEnabled(const std::string& vm_id) {
    return enabled_;
}

CBTResult HyperVCBTProvider::QueryChangedBlocks(
    const std::string& vm_id,
    const std::string& snapshot_id,
    uint64_t from_offset,
    uint64_t to_offset)
{
    CBTResult result;
    result.success = enabled_;
    result.changed_blocks = cached_changes_;
    
    for (const auto& block : cached_changes_) {
        result.total_changed_bytes += block.length;
    }
    
    return result;
}

bool HyperVCBTProvider::GetDiskStatistics(
    const std::string& vm_id,
    uint64_t& total_size,
    uint64_t& changed_size)
{
    total_size = 100ULL * 1024 * 1024 * 1024;
    changed_size = 0;
    
    for (const auto& block : cached_changes_) {
        changed_size += block.length;
    }
    
    return true;
}

VMwareCBTProvider::VMwareCBTProvider() = default;

bool VMwareCBTProvider::EnableCBT(const std::string& vm_id) {
    enabled_ = true;
    return true;
}

bool VMwareCBTProvider::DisableCBT(const std::string& vm_id) {
    enabled_ = false;
    return true;
}

bool VMwareCBTProvider::IsCBTEnabled(const std::string& vm_id) {
    return enabled_;
}

CBTResult VMwareCBTProvider::QueryChangedBlocks(
    const std::string& vm_id,
    const std::string& snapshot_id,
    uint64_t from_offset,
    uint64_t to_offset)
{
    CBTResult result;
    result.success = enabled_;
    result.changed_blocks = cached_changes_;
    
    for (const auto& block : cached_changes_) {
        result.total_changed_bytes += block.length;
    }
    
    return result;
}

bool VMwareCBTProvider::GetDiskStatistics(
    const std::string& vm_id,
    uint64_t& total_size,
    uint64_t& changed_size)
{
    total_size = 100ULL * 1024 * 1024 * 1024;
    changed_size = 0;
    
    for (const auto& block : cached_changes_) {
        changed_size += block.length;
    }
    
    return true;
}

KVMCBTProvider::KVMCBTProvider() = default;

bool KVMCBTProvider::EnableCBT(const std::string& vm_id) {
    enabled_ = true;
    return true;
}

bool KVMCBTProvider::DisableCBT(const std::string& vm_id) {
    enabled_ = false;
    return true;
}

bool KVMCBTProvider::IsCBTEnabled(const std::string& vm_id) {
    return enabled_;
}

CBTResult KVMCBTProvider::QueryChangedBlocks(
    const std::string& vm_id,
    const std::string& snapshot_id,
    uint64_t from_offset,
    uint64_t to_offset)
{
    CBTResult result;
    result.success = enabled_;
    result.changed_blocks = cached_changes_;
    
    for (const auto& block : cached_changes_) {
        result.total_changed_bytes += block.length;
    }
    
    return result;
}

bool KVMCBTProvider::GetDiskStatistics(
    const std::string& vm_id,
    uint64_t& total_size,
    uint64_t& changed_size)
{
    total_size = 100ULL * 1024 * 1024 * 1024;
    changed_size = 0;
    
    for (const auto& block : cached_changes_) {
        changed_size += block.length;
    }
    
    return true;
}

std::unique_ptr<ICBTProvider> CreateCBTProvider(const std::string& hypervisor_type) {
    if (hypervisor_type == "hyperv") {
        return std::make_unique<HyperVCBTProvider>();
    } else if (hypervisor_type == "vmware") {
        return std::make_unique<VMwareCBTProvider>();
    } else if (hypervisor_type == "kvm") {
        return std::make_unique<KVMCBTProvider>();
    }
    return nullptr;
}

}
