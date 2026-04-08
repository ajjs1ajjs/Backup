using Backup.Contracts;
using Grpc.Net.Client;
using Microsoft.AspNetCore.Mvc.Testing;
using Xunit;

namespace Backup.Server.IntegrationTests;

public class AgentGrpcIntegrationTests : IClassFixture<WebApplicationFactory<Program>>
{
    private readonly WebApplicationFactory<Program> _factory;

    public AgentGrpcIntegrationTests(WebApplicationFactory<Program> factory)
    {
        _factory = factory;
    }

    [Fact]
    public async Task Agent_CanRegister_Successfully()
    {
        // Arrange
        var client = _factory.CreateDefaultClient();
        var channel = GrpcChannel.ForAddress(client.BaseAddress!, new GrpcChannelOptions
        {
            HttpClient = _factory.CreateClient()
        });
        var grpcClient = new AgentService.AgentServiceClient(channel);

        var request = new AgentRegistrationRequest
        {
            AgentId = "test-agent-001",
            Hostname = "test-host",
            OsType = "linux",
            AgentVersion = "1.0.0",
            AgentType = AgentType.AgentTypeHyperv
        };

        // Act
        var response = await grpcClient.RegisterAsync(request);

        // Assert
        Assert.True(response.Success);
        Assert.Equal("Registration successful", response.Message);
        Assert.True(response.AssignedAgentId > 0);
    }

    [Fact]
    public async Task Agent_CanSendHeartbeat_AndReceivePing()
    {
        // Arrange
        var client = _factory.CreateClient();
        var channel = GrpcChannel.ForAddress(client.BaseAddress!, new GrpcChannelOptions
        {
            HttpClient = client
        });
        var grpcClient = new AgentService.AgentServiceClient(channel);

        // Спершу реєструємося
        var regResponse = await grpcClient.RegisterAsync(new AgentRegistrationRequest
        {
            AgentId = "heartbeat-agent",
            Hostname = "heartbeat-host",
            AgentType = AgentType.AgentTypeVmware
        });

        // Act
        using var call = grpcClient.Heartbeat();
        
        // Відправляємо серцебиття
        await call.RequestStream.WriteAsync(new AgentHeartbeat
        {
            AgentId = regResponse.AssignedAgentId,
            Status = AgentStatus.AgentStatusIdle
        });

        // Чекаємо відповідь від сервера
        Assert.True(await call.ResponseStream.MoveNext(CancellationToken.None));
        var command = call.ResponseStream.Current;

        // Assert
        Assert.NotNull(command.Ping);
        Assert.True(command.Ping.Timestamp > 0);

        await call.RequestStream.CompleteAsync();
    }
}
