#pragma once

#include <string>
#include <vector>
#include <cstdint>
#include <memory>
#include <chrono>
#include <optional>

namespace backup {

struct DatabaseInfo {
    std::string id;
    std::string name;
    std::string type; // mssql, postgresql, oracle
    std::string host;
    int port = 0;
    std::string username;
    uint64_t size_bytes = 0;
    bool is_available = true;
};

struct BackupMetadata {
    std::string backup_id;
    std::string database_name;
    std::string backup_type; // full, differential, log
    std::string file_path;
    uint64_t size_bytes = 0;
    std::chrono::system_clock::time_point start_time;
    std::chrono::system_clock::time_point end_time;
    bool success = false;
    std::string error_message;
};

struct RestorePoint {
    std::string backup_id;
    std::chrono::system_clock::time_point timestamp;
    std::string backup_type;
    uint64_t size_bytes = 0;
    bool is_available = true;
};

class IDatabaseAgent {
public:
    virtual ~IDatabaseAgent() = default;
    
    virtual bool Connect(const std::string& connection_string) = 0;
    virtual void Disconnect() = 0;
    virtual bool IsConnected() const = 0;
    
    virtual std::vector<DatabaseInfo> ListDatabases() = 0;
    virtual DatabaseInfo* GetDatabase(const std::string& db_name) = 0;
    
    virtual BackupMetadata CreateBackup(
        const std::string& database_name,
        const std::string& backup_type,
        const std::string& destination_path) = 0;
    
    virtual bool RestoreBackup(
        const std::string& database_name,
        const std::string& backup_id,
        const std::string& restore_point) = 0;
    
    virtual std::vector<RestorePoint> ListBackups(const std::string& database_name) = 0;
    
    virtual bool VerifyBackup(const std::string& backup_id) = 0;
};

class MSSQLAgent : public IDatabaseAgent {
public:
    MSSQLAgent();
    ~MSSQLAgent() override;
    
    bool Connect(const std::string& connection_string) override;
    void Disconnect() override;
    bool IsConnected() const override;
    
    std::vector<DatabaseInfo> ListDatabases() override;
    DatabaseInfo* GetDatabase(const std::string& db_name) override;
    
    BackupMetadata CreateBackup(
        const std::string& database_name,
        const std::string& backup_type,
        const std::string& destination_path) override;
    
    bool RestoreBackup(
        const std::string& database_name,
        const std::string& backup_id,
        const std::string& restore_point) override;
    
    std::vector<RestorePoint> ListBackups(const std::string& database_name) override;
    
    bool VerifyBackup(const std::string& backup_id) override;

private:
    bool connected_ = false;
    std::string connection_string_;
    std::vector<DatabaseInfo> databases_;
};

class PostgreSQLAgent : public IDatabaseAgent {
public:
    PostgreSQLAgent();
    ~PostgreSQLAgent() override;
    
    bool Connect(const std::string& connection_string) override;
    void Disconnect() override;
    bool IsConnected() const override;
    
    std::vector<DatabaseInfo> ListDatabases() override;
    DatabaseInfo* GetDatabase(const std::string& db_name) override;
    
    BackupMetadata CreateBackup(
        const std::string& database_name,
        const std::string& backup_type,
        const std::string& destination_path) override;
    
    bool RestoreBackup(
        const std::string& database_name,
        const std::string& backup_id,
        const std::string& restore_point) override;
    
    std::vector<RestorePoint> ListBackups(const std::string& database_name) override;
    
    bool VerifyBackup(const std::string& backup_id) override;

private:
    bool connected_ = false;
    std::string connection_string_;
    std::vector<DatabaseInfo> databases_;
};

class OracleAgent : public IDatabaseAgent {
public:
    OracleAgent();
    ~OracleAgent() override;
    
    bool Connect(const std::string& connection_string) override;
    void Disconnect() override;
    bool IsConnected() const override;
    
    std::vector<DatabaseInfo> ListDatabases() override;
    DatabaseInfo* GetDatabase(const std::string& db_name) override;
    
    BackupMetadata CreateBackup(
        const std::string& database_name,
        const std::string& backup_type,
        const std::string& destination_path) override;
    
    bool RestoreBackup(
        const std::string& database_name,
        const std::string& backup_id,
        const std::string& restore_point) override;
    
    std::vector<RestorePoint> ListBackups(const std::string& database_name) override;
    
    bool VerifyBackup(const std::string& backup_id) override;

private:
    bool connected_ = false;
    std::string connection_string_;
    std::vector<DatabaseInfo> databases_;
};

std::unique_ptr<IDatabaseAgent> CreateDatabaseAgent(const std::string& db_type);

}
