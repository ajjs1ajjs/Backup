using System.Threading.Tasks;
using NovaBackup.GUI.ViewModels;
using NovaBackup.GUI.Services;
using NovaBackup.GUI.Models;
using System.Collections.ObjectModel;
using Xunit;

namespace NovaBackup.UI.Tests
{
    public class StubApiClient : IApiClient
    {
        public Task<System.Collections.Generic.List<InfrastructureNode>> GetInfrastructureTreeAsync() => Task.FromResult(new System.Collections.Generic.List<InfrastructureNode>());
        public Task<System.Collections.Generic.List<RepositoryModel>> GetRepositoriesAsync() => Task.FromResult(new System.Collections.Generic.List<RepositoryModel>());
        public Task<System.Collections.Generic.List<JobModel>> GetJobsAsync() => Task.FromResult(new System.Collections.Generic.List<JobModel>());
        public Task<bool> CreateJobAsync(JobModel job) => Task.FromResult(true);
        public Task<bool> CreateJobAsync(JobModel job, string schedule) => Task.FromResult(true);
        public Task<System.Collections.Generic.List<RestorePointModel>> GetRestorePointsAsync(string jobId) => Task.FromResult(new System.Collections.Generic.List<RestorePointModel>());
        public Task<bool> StartInstantRecoveryAsync(string rpId, string vmName) => Task.FromResult(true);
        // Additional members omitted for brevity if IApiClient contains more members
    }

    public class JobWizardViewModelTests
    {
        [Fact]
        public async Task Finish_WithValidData_CallsApiAndSetsStatus()
        {
            var api = new StubApiClient();
            var vm = new JobWizardViewModel(api) { JobName = "Job1", EnableGuestProcessing = false };
            // set required selections for finish
            vm.SelectedSource = new InfrastructureNode { Name = "Src" };
            vm.SelectedRepository = new RepositoryModel { Name = "Repo" };
            await vm.Finish(null);
            Assert.Contains("Job", vm.StatusMessage);
        }

        [Fact]
        public void NextStep_Increments()
        {
            var api = new StubApiClient();
            var vm = new JobWizardViewModel(api);
            int initial = vm.CurrentStep;
            vm.NextStepCommand.Execute(null);
            Assert.Equal(initial + 1, vm.CurrentStep);
        }

        [Fact]
        public void ScheduleDisplay_Updates_When_Type_And_Time_Change()
        {
            var api = new StubApiClient();
            var vm = new JobWizardViewModel(api);
            vm.ScheduleType = "Weekly";
            vm.ScheduleTime = "23:30";
            Assert.Equal("Weekly 23:30", vm.ScheduleDisplay);
        }
    }
}
