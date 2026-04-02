#pragma once

#include <string>
#include <vector>
#include <cstdint>
#include <memory>
#include <functional>

namespace backup {

struct KVMVM {
    std::string vm_id;
    std::string name;
    std::string host;
    std::string ip_address;
    std::string os_type;
    uint64_t memory_mb = 0;
    int32_t cpu_cores = 0;
    std::vector<std::string> disks;
    std::string status;
};

struct KVMSnapshot {
    std::string id;
    std::string name;
    std::string created;
};

class IKVMAgent {
public:
    virtual ~IKVMAgent() = default;
    
    virtual bool Connect(const std::string& host, const std::string& username, const std::string& password) = 0;
    virtual void Disconnect() = 0;
    virtual bool IsConnected() const = 0;
    
    virtual std::vector<KVMVM> ListVMs() = 0;
    virtual KVMVM* GetVM(const std::string& vm_id) = 0;
    
    virtual bool CreateSnapshot(const std::string& vm_id, const std::string& name) = 0;
    virtual bool DeleteSnapshot(const std::string& vm_id, const std::string& snapshot_id) = 0;
    virtual std::vector<KVMSnapshot> ListSnapshots(const std::string& vm_id) = 0;
    virtual bool RevertToSnapshot(const std::string& vm_id, const std::string& snapshot_id) = 0;
    
    virtual std::string GetDiskPath(const std::string& vm_id, int disk_number) = 0;
    virtual uint64_t GetChangedBlocks(const std::string& vm_id, const std::string& snapshot_id) = 0;
};

class KVMAgent : public IKVMAgent {
public:
    KVMAgent();
    ~KVMAgent() override;
    
    bool Connect(const std::string& host, const std::string& username, const std::string& password) override;
    void Disconnect() override;
    bool IsConnected() const override;
    
    std::vector<KVMVM> ListVMs() override;
    KVMVM* GetVM(const std::string& vm_id) override;
    
    bool CreateSnapshot(const std::string& vm_id, const std::string& name) override;
    bool DeleteSnapshot(const std::string& vm_id, const std::string& snapshot_id) override;
    std::vector<KVMSnapshot> ListSnapshots(const std::string& vm_id) override;
    bool RevertToSnapshot(const std::string& vm_id, const std::string& snapshot_id) override;
    
    std::string GetDiskPath(const std::string& vm_id, int disk_number) override;
    uint64_t GetChangedBlocks(const std::string& vm_id, const std::string& snapshot_id) override;

private:
    bool connected_ = false;
    std::string host_;
    std::vector<KVMVM> cached_vms_;
};

std::unique_ptr<IKVMAgent> CreateKVMAgent();

}
