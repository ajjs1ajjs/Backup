#pragma once

#include <string>
#include <memory>
#include <grpcpp/grpcpp.h>
#include "backup.grpc.pb.h"

class AgentClient {
public:
    AgentClient(std::shared_ptr<grpc::Channel> channel);
    ~AgentClient();

    bool RegisterAgent(const std::string& hostname, const std::string& os_type,
                      const std::string& agent_version, const std::string& agent_type);
    
    bool SendHeartbeat(int64_t agent_id, const std::string& status);
    
    bool StartBackup(const std::string& job_id, const std::string& backup_type,
                    const std::string& source, const std::string& destination);
    
    bool StopBackup(const std::string& job_id);
    
    void HeartbeatLoop();

private:
    std::unique_ptr<backup::AgentService::Stub> stub_;
    std::unique_ptr<backup::FileTransferService::Stub> transfer_stub_;
    std::string server_address_;
    std::string agent_token_;
};

class Config {
public:
    static Config& Instance();
    
    std::string server_address() const { return server_address_; }
    void set_server_address(const std::string& addr) { server_address_ = addr; }
    
    std::string agent_token() const { return agent_token_; }
    void set_agent_token(const std::string& token) { agent_token_ = token; }
    
    std::string agent_type() const { return agent_type_; }
    void set_agent_type(const std::string& type) { agent_type_ = type; }
    
    std::string log_level() const { return log_level_; }
    void set_log_level(const std::string& level) { log_level_ = level; }
    
    bool LoadFromFile(const std::string& path);
    bool LoadFromEnv();
    void SaveToFile(const std::string& path) const;

private:
    Config() = default;
    std::string server_address_;
    std::string agent_token_;
    std::string agent_type_;
    std::string log_level_ = "info";
};
