#pragma once

#include <string>
#include <vector>
#include <cstdint>
#include <memory>
#include <functional>

namespace backup {

struct VMwareVM {
    std::string vm_id;
    std::string name;
    std::string host;
    std::string ip_address;
    std::string guest_os;
    uint64_t memory_mb = 0;
    int32_t cpu_cores = 0;
    std::vector<std::string> disks;
    std::string status;
};

struct VMwareSnapshot {
    std::string id;
    std::string name;
    std::string created;
    bool is_current = false;
};

class IVMwareAgent {
public:
    virtual ~IVMwareAgent() = default;
    
    virtual bool Connect(const std::string& host, const std::string& username, const std::string& password) = 0;
    virtual void Disconnect() = 0;
    virtual bool IsConnected() const = 0;
    
    virtual std::vector<VMwareVM> ListVMs() = 0;
    virtual VMwareVM* GetVM(const std::string& vm_id) = 0;
    
    virtual bool CreateSnapshot(const std::string& vm_id, const std::string& name) = 0;
    virtual bool DeleteSnapshot(const std::string& vm_id, const std::string& snapshot_id) = 0;
    virtual std::vector<VMwareSnapshot> ListSnapshots(const std::string& vm_id) = 0;
    virtual bool RevertToSnapshot(const std::string& vm_id, const std::string& snapshot_id) = 0;
    
    virtual std::string GetDiskPath(const std::string& vm_id, int disk_number) = 0;
    virtual uint64_t GetChangedBlocks(const std::string& vm_id, const std::string& snapshot_id) = 0;
};

class VMwareAgent : public IVMwareAgent {
public:
    VMwareAgent();
    ~VMwareAgent() override;
    
    bool Connect(const std::string& host, const std::string& username, const std::string& password) override;
    void Disconnect() override;
    bool IsConnected() const override;
    
    std::vector<VMwareVM> ListVMs() override;
    VMwareVM* GetVM(const std::string& vm_id) override;
    
    bool CreateSnapshot(const std::string& vm_id, const std::string& name) override;
    bool DeleteSnapshot(const std::string& vm_id, const std::string& snapshot_id) override;
    std::vector<VMwareSnapshot> ListSnapshots(const std::string& vm_id) override;
    bool RevertToSnapshot(const std::string& vm_id, const std::string& snapshot_id) override;
    
    std::string GetDiskPath(const std::string& vm_id, int disk_number) override;
    uint64_t GetChangedBlocks(const std::string& vm_id, const std::string& snapshot_id) override;

private:
    bool connected_ = false;
    std::string host_;
    std::string vddk_path_;
    std::vector<VMwareVM> cached_vms_;
};

std::unique_ptr<IVMwareAgent> CreateVMwareAgent();

}
