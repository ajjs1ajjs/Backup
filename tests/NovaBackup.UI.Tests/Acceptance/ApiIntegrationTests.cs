using System;
using System.Net;
using System.Net.Http;
using System.Threading;
using System.Threading.Tasks;
using System.Collections.Generic;
using NovaBackup.GUI.Models;
using NovaBackup.GUI.Services;
using Xunit;

namespace NovaBackup.UI.Tests.Acceptance
{
    // Lightweight fake backend to exercise ApiClient against common endpoints
    public class ApiIntegrationTests
    {
        private class FakeBackendHandler : HttpMessageHandler
        {
            public HttpRequestMessage LastRequest { get; private set; }
            protected override Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken cancellationToken)
            {
                LastRequest = request;
                // Simple routing based on path
                if (request.Method == HttpMethod.Get && request.RequestUri.AbsolutePath.Contains("infrastructure/tree"))
                {
                    var json = "[{\"id\":\"src1\",\"name\":\"LocalDisk\" ,\"type\":\"disk\"}]";
                    return Task.FromResult(new HttpResponseMessage(HttpStatusCode.OK)
                    { Content = new StringContent(json, System.Text.Encoding.UTF8, "application/json") });
                }
                if (request.Method == HttpMethod.Get && request.RequestUri.AbsolutePath.Contains("storage/repositories"))
                {
                    var json = "[{\"name\":\"Repo1\"}]";
                    return Task.FromResult(new HttpResponseMessage(HttpStatusCode.OK)
                    { Content = new StringContent(json, System.Text.Encoding.UTF8, "application/json") });
                }
                if (request.Method == HttpMethod.Post && request.RequestUri.AbsolutePath.Contains("jobs"))
                {
                    return Task.FromResult(new HttpResponseMessage(HttpStatusCode.OK)
                    { Content = new StringContent("{\"ok\":true}", System.Text.Encoding.UTF8, "application/json") });
                }
                return Task.FromResult(new HttpResponseMessage(HttpStatusCode.NotFound));
            }
        }

        [Fact]
        public async Task ApiClient_Endpoints_Work_With_FakeBackend()
        {
            var handler = new FakeBackendHandler();
            var httpClient = new HttpClient(handler);
            var api = new ApiClient(httpClient);

            // Get infrastructure tree
            var infra = await api.GetInfrastructureTreeAsync();
            Assert.NotNull(infra);

            // Get repositories
            var repos = await api.GetRepositoriesAsync();
            Assert.NotNull(repos);

            // Create a job via explicit schedule
            var job = new JobModel { Name = "IntegrationJob", Platform = "VMware", RetentionDays = 7, EnableGuestProcessing = false, GuestCredentialsId = "" };
            var ok = await api.CreateJobAsync(job, "Daily 22:00");
            Assert.True(ok);
        }
    }
}
