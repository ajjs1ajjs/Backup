#include "agent_client.h"
#include <iostream>
#include <thread>
#include <chrono>

AgentClient::AgentClient(std::shared_ptr<grpc::Channel> channel)
    : stub_(backup::AgentService::NewStub(channel)) {
}

AgentClient::~AgentClient() {
}

bool AgentClient::RegisterAgent(const std::string& hostname, const std::string& os_type,
                               const std::string& agent_version, const std::string& agent_type) {
    backup::AgentRegistrationRequest request;
    request.set_agent_id(hostname); // Use hostname as simple ID for now
    request.set_hostname(hostname);
    request.set_os_type(os_type);
    request.set_agent_version(agent_version);
    
    if (agent_type == "hyperv") request.set_agent_type(backup::AgentType::Hyperv);
    else if (agent_type == "vmware") request.set_agent_type(backup::AgentType::Vmware);
    else if (agent_type == "kvm") request.set_agent_type(backup::AgentType::Kvm);

    backup::AgentRegistrationResponse response;
    grpc::ClientContext context;
    
    // Add registration token from config/env
    std::string reg_token = "test-token-123"; // Should be from Config
    context.AddMetadata("x-registration-token", reg_token);

    grpc::Status status = stub_->Register(&context, request, &response);

    if (status.ok() && response.success()) {
        auto it = context.GetServerTrailingMetadata().find("x-agent-token");
        if (it != context.GetServerTrailingMetadata().end()) {
            agent_token_ = std::string(it->second.data(), it->second.length());
        }
        return true;
    }

    std::cerr << "Registration failed: " << status.error_message() << std::endl;
    return false;
}

void AgentClient::HeartbeatLoop() {
    while (true) {
        grpc::ClientContext context;
        context.AddMetadata("x-agent-token", agent_token_);

        auto stream = stub_->Heartbeat(&context);

        backup::AgentHeartbeat heartbeat;
        heartbeat.set_status(backup::AgentStatus::Idle);
        
        if (!stream->Write(heartbeat)) {
            std::cerr << "Failed to write heartbeat" << std::endl;
            std::this_thread::sleep_for(std::chrono::seconds(5));
            continue;
        }

        backup::ServerCommand command;
        while (stream->Read(&command)) {
            if (command.has_ping()) {
                std::cout << "Received Ping from server" << std::endl;
            } else if (command.has_start_backup()) {
                std::cout << "Received StartBackup command" << std::endl;
            }
            
            // Send next heartbeat
            std::this_thread::sleep_for(std::chrono::seconds(30));
            heartbeat.set_status(backup::AgentStatus::Idle);
            if (!stream->Write(heartbeat)) break;
        }

        std::cerr << "Heartbeat stream closed. Reconnecting..." << std::endl;
        std::this_thread::sleep_for(std::chrono::seconds(5));
    }
}
