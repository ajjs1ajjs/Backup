using System;
using System.Net;
using System.Net.Http;
using System.Text;
using System.Text.Json;
using System.Threading;
using System.Threading.Tasks;
using NovaBackup.GUI.Services;
using NovaBackup.GUI.Models;
using Xunit;

namespace NovaBackup.UI.Tests
{
    public class ApiClientPayloadTests
    {
        private class InspectingHandler : HttpMessageHandler
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

        [Fact]
        public async Task CreateJobAsync_WithExplicitSchedule_SendsExpectedFields()
        {
            var handler = new InspectingHandler(new HttpResponseMessage(HttpStatusCode.OK)
            {
                Content = new StringContent("{\"ok\":true}", Encoding.UTF8, "application/json")
            });
            var httpClient = new HttpClient(handler);
            var api = new ApiClient(httpClient);
            var job = new JobModel
            {
                Name = "JobX",
                Platform = "VMware",
                RetentionDays = 5,
                EnableGuestProcessing = true,
                GuestCredentialsId = "cred-123"
            };

            var ok = await api.CreateJobAsync(job, "Daily 22:00");
            Assert.True(ok);
            var req = handler.LastRequest;
            Assert.NotNull(req);
            var body = await req.Content.ReadAsStringAsync();
            using var doc = JsonDocument.Parse(body);
            var root = doc.RootElement;
            Assert.Equal(job.Name, root.GetProperty("name").GetString());
            Assert.Equal("Daily 22:00", root.GetProperty("schedule").GetString());
            Assert.Equal(5, root.GetProperty("retention_days").GetInt32());
            Assert.True(root.GetProperty("guest_processing").GetBoolean());
            Assert.Equal("cred-123", root.GetProperty("guest_credentials_id").GetString());
        }

        [Fact]
        public async Task CreateJobAsync_WithoutExplicitSchedule_SendsDefaultSchedule()
        {
            var handler = new InspectingHandler(new HttpResponseMessage(HttpStatusCode.OK)
            {
                Content = new StringContent("{\"ok\":true}", Encoding.UTF8, "application/json")
            });
            var httpClient = new HttpClient(handler);
            var api = new ApiClient(httpClient);
            var job = new JobModel
            {
                Name = "JobY",
                Platform = "Hyper-V",
                RetentionDays = 14,
                EnableGuestProcessing = false,
                GuestCredentialsId = string.Empty
            };

            var ok = await api.CreateJobAsync(job);
            Assert.True(ok);
            var body = await handler.LastRequest.Content.ReadAsStringAsync();
            using var doc = JsonDocument.Parse(body);
            var root = doc.RootElement;
            Assert.True(root.TryGetProperty("schedule", out var _));
        }
    }
}
