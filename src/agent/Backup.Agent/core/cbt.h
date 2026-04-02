#pragma once

#include <string>
#include <vector>
#include <cstdint>
#include <memory>

namespace backup {

struct ChangedBlock {
    uint64_t offset = 0;
    uint64_t length = 0;
};

struct CBTResult {
    std::vector<ChangedBlock> changed_blocks;
    uint64_t total_changed_bytes = 0;
    bool success = false;
    std::string error_message;
};

class ICBTProvider {
public:
    virtual ~ICBTProvider() = default;
    
    virtual bool EnableCBT(const std::string& vm_id) = 0;
    virtual bool DisableCBT(const std::string& vm_id) = 0;
    virtual bool IsCBTEnabled(const std::string& vm_id) = 0;
    
    virtual CBTResult QueryChangedBlocks(
        const std::string& vm_id,
        const std::string& snapshot_id,
        uint64_t from_offset,
        uint64_t to_offset) = 0;
    
    virtual bool GetDiskStatistics(
        const std::string& vm_id,
        uint64_t& total_size,
        uint64_t& changed_size) = 0;
};

class HyperVCBTProvider : public ICBTProvider {
public:
    HyperVCBTProvider();
    
    bool EnableCBT(const std::string& vm_id) override;
    bool DisableCBT(const std::string& vm_id) override;
    bool IsCBTEnabled(const std::string& vm_id) override;
    
    CBTResult QueryChangedBlocks(
        const std::string& vm_id,
        const std::string& snapshot_id,
        uint64_t from_offset,
        uint64_t to_offset) override;
    
    bool GetDiskStatistics(
        const std::string& vm_id,
        uint64_t& total_size,
        uint64_t& changed_size) override;

private:
    std::vector<ChangedBlock> cached_changes_;
    bool enabled_ = false;
};

class VMwareCBTProvider : public ICBTProvider {
public:
    VMwareCBTProvider();
    
    bool EnableCBT(const std::string& vm_id) override;
    bool DisableCBT(const std::string& vm_id) override;
    bool IsCBTEnabled(const std::string& vm_id) override;
    
    CBTResult QueryChangedBlocks(
        const std::string& vm_id,
        const std::string& snapshot_id,
        uint64_t from_offset,
        uint64_t to_offset) override;
    
    bool GetDiskStatistics(
        const std::string& vm_id,
        uint64_t& total_size,
        uint64_t& changed_size) override;

private:
    std::vector<ChangedBlock> cached_changes_;
    bool enabled_ = false;
};

class KVMCBTProvider : public ICBTProvider {
public:
    KVMCBTProvider();
    
    bool EnableCBT(const std::string& vm_id) override;
    bool DisableCBT(const std::string& vm_id) override;
    bool IsCBTEnabled(const std::string& vm_id) override;
    
    CBTResult QueryChangedBlocks(
        const std::string& vm_id,
        const std::string& snapshot_id,
        uint64_t from_offset,
        uint64_t to_offset) override;
    
    bool GetDiskStatistics(
        const std::string& vm_id,
        uint64_t& total_size,
        uint64_t& changed_size) override;

private:
    std::vector<ChangedBlock> cached_changes_;
    bool enabled_ = false;
};

std::unique_ptr<ICBTProvider> CreateCBTProvider(const std::string& hypervisor_type);

}
