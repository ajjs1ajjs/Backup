using System.Threading.Tasks;
using NovaBackup.GUI.ViewModels;
using NovaBackup.GUI.Services;
using NovaBackup.GUI.Models;
using Xunit;

namespace NovaBackup.UI.Tests
{
    public class ScheduleDisplayTests
    {
        private class DummyApi : IApiClient
        {
            public Task<System.Collections.Generic.List<InfrastructureNode>> GetInfrastructureTreeAsync() => Task.FromResult(new System.Collections.Generic.List<InfrastructureNode>());
            public Task<System.Collections.Generic.List<RepositoryModel>> GetRepositoriesAsync() => Task.FromResult(new System.Collections.Generic.List<RepositoryModel>());
            public Task<System.Collections.Generic.List<JobModel>> GetJobsAsync() => Task.FromResult(new System.Collections.Generic.List<JobModel>>());
            public Task<bool> CreateJobAsync(JobModel job) => Task.FromResult(true);
            public Task<bool> CreateJobAsync(JobModel job, string schedule) => Task.FromResult(true);
            public Task<System.Collections.Generic.List<RestorePointModel>> GetRestorePointsAsync(string jobId) => Task.FromResult(new System.Collections.Generic.List<RestorePointModel>());
            public Task<bool> StartInstantRecoveryAsync(string rpId, string vmName) => Task.FromResult(true);
        }

        [Fact]
        public void ScheduleDisplay_Updates_On_Type_Or_Time_Change()
        {
            var vm = new JobWizardViewModel(new DummyApi());
            vm.ScheduleType = "Daily";
            vm.ScheduleTime = "22:00";
            Assert.Equal("Daily 22:00", vm.ScheduleDisplay);

            vm.ScheduleType = "Weekly";
            vm.ScheduleTime = "23:30";
            Assert.Equal("Weekly 23:30", vm.ScheduleDisplay);
        }
    }
}
