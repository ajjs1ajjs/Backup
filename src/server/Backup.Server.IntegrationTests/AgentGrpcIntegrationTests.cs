using Backup.Contracts;
using Grpc.Core;
using Grpc.Net.Client;
using Microsoft.AspNetCore.Mvc.Testing;
using Xunit;

namespace Backup.Server.IntegrationTests;

public class AgentGrpcIntegrationTests : IClassFixture<IntegrationTestWebApplicationFactory>
{
    private readonly IntegrationTestWebApplicationFactory _factory;

    public AgentGrpcIntegrationTests(IntegrationTestWebApplicationFactory factory)
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
            AgentType = AgentType.Hyperv
        };

        // Act
        var headers = new Metadata { { "x-registration-token", "test-token-123" } };
        var call = grpcClient.RegisterAsync(request, headers);
        var response = await call.ResponseAsync;

        // Assert
        if (!response.Success)
        {
            throw new Exception($"Registration failed: {response.Message}");
        }
        Assert.True(response.Success);
        Assert.Equal("Registration successful", response.Message);
        Assert.True(response.AssignedAgentId > 0);
        
        var authToken = (await call.ResponseHeadersAsync).GetValue("x-agent-token");
        Assert.False(string.IsNullOrEmpty(authToken));
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
        var headers = new Metadata { { "x-registration-token", "test-token-123" } };
        var regCall = grpcClient.RegisterAsync(new AgentRegistrationRequest
        {
            AgentId = "heartbeat-agent",
            Hostname = "heartbeat-host",
            AgentType = AgentType.Vmware
        }, headers);
        
        var regResponse = await regCall.ResponseAsync;
        if (!regResponse.Success) throw new Exception($"Reg failed: {regResponse.Message}");
        
        var authToken = (await regCall.ResponseHeadersAsync).GetValue("x-agent-token");
        if (string.IsNullOrEmpty(authToken)) throw new Exception("Auth token not received in headers");

        Console.WriteLine($"Agent registered with ID {regResponse.AssignedAgentId} and token {authToken}");

        // Act - Try without headers first to see if it reaches the server
        Console.WriteLine("Starting heartbeat call...");
        using var call = grpcClient.Heartbeat();
        
        // Відправляємо серцебиття
        Console.WriteLine("Sending first heartbeat message...");
        await call.RequestStream.WriteAsync(new AgentHeartbeat
        {
            AgentId = regResponse.AssignedAgentId,
            Status = AgentStatus.Idle
        });
        Console.WriteLine("First heartbeat message sent.");

        // Чекаємо відповідь від сервера
        Console.WriteLine("Waiting for server response...");
        var moveNextTask = call.ResponseStream.MoveNext(CancellationToken.None);
        if (await Task.WhenAny(moveNextTask, Task.Delay(10000)) != moveNextTask)
        {
            throw new Exception("Timed out waiting for heartbeat response from server");
        }
        
        Assert.True(await moveNextTask);
        var command = call.ResponseStream.Current;

        // Assert
        Assert.NotNull(command.Ping);
        Assert.True(command.Ping.Timestamp > 0);

        await call.RequestStream.CompleteAsync();
    }
}
