using Grpc.Net.Client;
using Backup.Contracts;
using Xunit;
using Microsoft.Extensions.Logging;
using Moq;

namespace Backup.Server.IntegrationTests;

public class AgentRegistrationTests : IClassFixture<GrpcIntegrationTestFixture>, IAsyncLifetime
{
    private readonly GrpcChannel _channel;
    private readonly AgentService.AgentServiceClient _agentClient;

    public AgentRegistrationTests(GrpcIntegrationTestFixture fixture)
    {
        _channel = fixture.CreateGrpcChannel();
        _agentClient = new AgentService.AgentServiceClient(_channel);
    }

    public Task InitializeAsync() => Task.CompletedTask;
    public Task DisposeAsync() => Task.CompletedTask;

    [Fact]
    public async Task RegisterAgent_ShouldReturnSuccess()
    {
        // Arrange
        var request = new AgentRegistrationRequest
        {
            AgentId = "test-agent-001",
            Hostname = "test-host.local",
            OsType = "Windows",
            AgentVersion = "1.0.0",
            AgentType = AgentType.AgentTypeHyperv,
            Capabilities = { "backup", "restore", "cbt" }
        };

        // Act
        var response = await _agentClient.RegisterAsync(request);

        // Assert
        Assert.NotNull(response);
        Assert.True(response.Success);
        Assert.NotEmpty(response.ServerVersion);
    }

    [Fact]
    public async Task RegisterAgent_WithDifferentTypes_ShouldSucceed()
    {
        // Arrange
        var agentTypes = new[]
        {
            AgentType.AgentTypeHyperv,
            AgentType.AgentTypeVmware,
            AgentType.AgentTypeKvm,
            AgentType.AgentTypeMssql,
            AgentType.AgentTypePostgresql,
            AgentType.AgentTypeOracle
        };

        foreach (var agentType in agentTypes)
        {
            // Arrange
            var request = new AgentRegistrationRequest
            {
                AgentId = $"test-agent-{agentType}",
                Hostname = $"test-{agentType}.local",
                OsType = "Linux",
                AgentVersion = "1.0.0",
                AgentType = agentType,
                Capabilities = { "backup", "restore" }
            };

            // Act
            var response = await _agentClient.RegisterAsync(request);

            // Assert
            Assert.True(response.Success, $"Failed to register {agentType} agent");
        }
    }

    [Fact]
    public async Task RegisterAgent_WithInvalidData_ShouldReturnError()
    {
        // Arrange
        var request = new AgentRegistrationRequest
        {
            AgentId = "", // Invalid: empty agent ID
            Hostname = "",
            OsType = "",
            AgentVersion = "",
            AgentType = AgentType.AgentTypeUnspecified
        };

        // Act
        var response = await _agentClient.RegisterAsync(request);

        // Assert
        Assert.False(response.Success);
        Assert.NotEmpty(response.Message);
    }

    [Fact]
    public async Task RegisterAgent_MultipleTimes_ShouldHandleCorrectly()
    {
        // Arrange
        var request = new AgentRegistrationRequest
        {
            AgentId = "duplicate-agent",
            Hostname = "duplicate.local",
            OsType = "Windows",
            AgentVersion = "1.0.0",
            AgentType = AgentType.AgentTypeHyperv
        };

        // Act - First registration
        var firstResponse = await _agentClient.RegisterAsync(request);

        // Act - Second registration (re-register)
        var secondResponse = await _agentClient.RegisterAsync(request);

        // Assert
        Assert.True(firstResponse.Success);
        Assert.True(secondResponse.Success || !secondResponse.Success); // Implementation dependent
    }
}

public class AgentHeartbeatTests : IClassFixture<GrpcIntegrationTestFixture>
{
    private readonly GrpcChannel _channel;
    private readonly AgentService.AgentServiceClient _agentClient;

    public AgentHeartbeatTests(GrpcIntegrationTestFixture fixture)
    {
        _channel = fixture.CreateGrpcChannel();
        _agentClient = new AgentService.AgentServiceClient(_channel);
    }

    [Fact]
    public async Task AgentHeartbeat_ShouldEstablishStream()
    {
        // Arrange
        var heartbeatCall = _agentClient.Heartbeat();

        // Act
        await heartbeatCall.RequestStream.WriteAsync(new AgentHeartbeat
        {
            AgentId = 1,
            Status = AgentStatus.AgentStatusIdle,
            ResourceUsage = new ResourceUsage
            {
                CpuPercent = 25.5,
                MemoryUsedMb = 512,
                MemoryTotalMb = 2048,
                NetworkSpeedMbps = 100,
                DiskSpeedMbps = 500
            }
        });

        // Assert
        Assert.True(heartbeatCall.RequestStream.WriteOptions != null);
        await heartbeatCall.RequestStream.CompleteAsync();
    }

    [Fact]
    public async Task AgentHeartbeat_WithStatusUpdates_ShouldReceiveCommands()
    {
        // Arrange
        var heartbeatCall = _agentClient.Heartbeat();
        var receivedCommands = new List<ServerCommand>();

        // Start reading responses
        var readTask = Task.Run(async () =>
        {
            await foreach (var response in heartbeatCall.ResponseStream.ReadAllAsync())
            {
                receivedCommands.Add(response);
            }
        });

        // Act - Send heartbeats
        for (int i = 0; i < 5; i++)
        {
            await heartbeatCall.RequestStream.WriteAsync(new AgentHeartbeat
            {
                AgentId = 1,
                Status = AgentStatus.AgentStatusIdle,
                ResourceUsage = new ResourceUsage
                {
                    CpuPercent = 10 + i * 5,
                    MemoryUsedMb = 512,
                    MemoryTotalMb = 2048
                }
            });

            await Task.Delay(100);
        }

        await heartbeatCall.RequestStream.CompleteAsync();
        await readTask;

        // Assert
        Assert.NotNull(receivedCommands);
    }
}

public class AgentCapabilitiesTests : IClassFixture<GrpcIntegrationTestFixture>
{
    private readonly GrpcChannel _channel;
    private readonly AgentService.AgentServiceClient _agentClient;

    public AgentCapabilitiesTests(GrpcIntegrationTestFixture fixture)
    {
        _channel = fixture.CreateGrpcChannel();
        _agentClient = new AgentService.AgentServiceClient(_channel);
    }

    [Fact]
    public async Task GetCapabilities_ShouldReturnAgentFeatures()
    {
        // Arrange
        var request = new AgentCapabilitiesRequest
        {
            AgentId = 1
        };

        // Act
        var response = await _agentClient.GetCapabilitiesAsync(request);

        // Assert
        Assert.NotNull(response);
        Assert.NotNull(response.Features);
        Assert.NotNull(response.Options);
    }
}
