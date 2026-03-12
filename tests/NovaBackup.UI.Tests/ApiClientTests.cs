using System;
using System.Net;
using System.Net.Http;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;
using NovaBackup.GUI.Services;
using Xunit;
using System.Collections.Generic;
using System.Linq;
using NovaBackup.GUI.Models;
using System.Threading;

namespace NovaBackup.UI.Tests
{
    // Simple delegating handler to inspect outgoing requests
    public class InspectingHandler : HttpMessageHandler
    {
        public HttpRequestMessage LastRequest { get; private set; }
        private readonly HttpResponseMessage _response;
        public InspectingHandler(HttpResponseMessage response)
        {
            _response = response;
        }
        protected override Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken cancellationToken)
        {
            LastRequest = request;
            return Task.FromResult(_response);
        }
    }

    public class ApiClientTests
    {
        [Fact]
        public async Task CreateJobAsync_SendsExpectedPayload_WithExplicitSchedule()
        {
            // arrange
            var payloadCapture = string.Empty;
            var handler = new InspectingHandler(new HttpResponseMessage(HttpStatusCode.OK)
            {
                Content = new StringContent("{\"ok\":true}", Encoding.UTF8, "application/json")
            });
            var httpClient = new HttpClient(handler);
            var api = new ApiClient(httpClient);
            var job = new JobModel
            {
                Name = "TestJob",
                Platform = "VMware",
                RetentionDays = 7,
                EnableGuestProcessing = true,
                GuestCredentialsId = "cred-id"
            };

            // act
            var result = await api.CreateJobAsync(job, "Daily 22:00");
            // assert
            Assert.True(result);
            var req = (handler as InspectingHandler).LastRequest;
            Assert.NotNull(req);
            var body = await req.Content.ReadAsStringAsync();
            using var doc = System.Text.Json.JsonDocument.Parse(body);
            var root = doc.RootElement;
            Assert.Equal("TestJob", root.GetProperty("name").GetString());
            Assert.Equal("Daily 22:00", root.GetProperty("schedule").GetString());
            Assert.Equal(7, root.GetProperty("retention_days").GetInt32());
            Assert.True(root.GetProperty("guest_processing").GetBoolean());
            Assert.Equal("cred-id", root.GetProperty("guest_credentials_id").GetString());
        }

        [Fact]
        public async Task CreateJobAsync_WithoutExplicitSchedule_SendsDefault()
        {
            var handler = new InspectingHandler(new HttpResponseMessage(HttpStatusCode.OK)
            {
                Content = new StringContent("{\"ok\":true}", Encoding.UTF8, "application/json")
            });
            var httpClient = new HttpClient(handler);
            var api = new ApiClient(httpClient);
            var job = new JobModel
            {
                Name = "DefaultJob",
                Platform = "Hyper-V",
                RetentionDays = 14,
                EnableGuestProcessing = false,
                GuestCredentialsId = ""
            };

            var result = await api.CreateJobAsync(job);
            Assert.True(result);
            var req = (handler as InspectingHandler).LastRequest;
            var body = await req.Content.ReadAsStringAsync();
            using var doc = System.Text.Json.JsonDocument.Parse(body);
            var root = doc.RootElement;
            Assert.True(root.TryGetProperty("schedule", out var _));
        }
    }
}
