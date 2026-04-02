#pragma once

#include <string>
#include <vector>
#include <memory>
#include <functional>
#include <cstdint>
#include <chrono>

namespace backup {

struct VirtualDisk {
    std::string id;
    std::string path;
    uint64_t size_bytes = 0;
    std::string format;
    bool is_differential = false;
    std::string parent_path;
};

struct VirtualMachine {
    std::string vm_id;
    std::string name;
    std::string hypervisor_type;
    std::string host;
    std::string ip_address;
    std::string os_type;
    uint64_t memory_mb = 0;
    int32_t cpu_cores = 0;
    std::vector<VirtualDisk> disks;
    std::string status;
};

struct Snapshot {
    std::string snapshot_id;
    std::string vm_id;
    std::string name;
    std::string created_at;
    bool is_current = false;
};

using VMCallback = std::function<void(const VirtualMachine&)>;
using SnapshotCallback = std::function<void(const Snapshot&)>;

class IHypervisorAgent {
public:
    virtual ~IHypervisorAgent() = default;
    
    virtual bool Connect(const std::string& host, const std::string& username, const std::string& password) = 0;
    virtual void Disconnect() = 0;
    virtual bool IsConnected() const = 0;
    
    virtual std::vector<VirtualMachine> ListVMs() = 0;
    virtual VirtualMachine* GetVM(const std::string& vm_id) = 0;
    
    virtual bool CreateSnapshot(const std::string& vm_id, const std::string& snapshot_name) = 0;
    virtual bool DeleteSnapshot(const std::string& vm_id, const std::string& snapshot_id) = 0;
    virtual std::vector<Snapshot> ListSnapshots(const std::string& vm_id) = 0;
    virtual bool RevertToSnapshot(const std::string& vm_id, const std::string& snapshot_id) = 0;
    
    virtual std::string GetDiskPath(const std::string& vm_id, int32_t disk_number) = 0;
    virtual uint64_t GetChangedBlocks(const std::string& vm_id, const std::string& snapshot_id) = 0;
};

class HyperVAgent : public IHypervisorAgent {
public:
    HyperVAgent();
    ~HyperVAgent() override;
    
    bool Connect(const std::string& host, const std::string& username, const std::string& password) override;
    void Disconnect() override;
    bool IsConnected() const override;
    
    std::vector<VirtualMachine> ListVMs() override;
    VirtualMachine* GetVM(const std::string& vm_id) override;
    
    bool CreateSnapshot(const std::string& vm_id, const std::string& snapshot_name) override;
    bool DeleteSnapshot(const std::string& vm_id, const std::string& snapshot_id) override;
    std::vector<Snapshot> ListSnapshots(const std::string& vm_id) override;
    bool RevertToSnapshot(const std::string& vm_id, const std::string& snapshot_id) override;
    
    std::string GetDiskPath(const std::string& vm_id, int32_t disk_number) override;
    uint64_t GetChangedBlocks(const std::string& vm_id, const std::string& snapshot_id) override;

private:
    bool connected_;
    std::string host_;
    std::vector<VirtualMachine> cached_vms_;
};

std::unique_ptr<IHypervisorAgent> CreateHyperVAgent();

}
