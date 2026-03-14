using System;
using System.Collections.Generic;
using System.Net.Http;
using System.Threading.Tasks;
using Moq;
using NovaBackup.GUI.Services;
using NovaBackup.GUI.ViewModels;
using Xunit;

namespace NovaBackup.WPF.Tests.ViewModels;

public class ReplicationViewModelTests
{
    private readonly Mock<IApiClient> _mockApiClient;
    private readonly ReplicationViewModel _viewModel;

    public ReplicationViewModelTests()
    {
        _mockApiClient = new Mock<IApiClient>();
        _viewModel = new ReplicationViewModel(_mockApiClient.Object);
    }

    [Fact]
    public void Constructor_InitializesWithEmptyCollection()
    {
        Assert.NotNull(_viewModel.ReplicationJobs);
        Assert.Empty(_viewModel.ReplicationJobs);
        Assert.False(_viewModel.IsLoading);
    }

    [Fact]
    public async Task LoadDataAsync_WhenSuccess_UpdatesCollection()
    {
        var expectedJobs = new List<ReplicationJobModel>
        {
            new ReplicationJobModel { Id = "1", Name = "VM Replication 1", Status = "Running", Progress = 50 },
            new ReplicationJobModel { Id = "2", Name = "VM Replication 2", Status = "Completed", Progress = 100 }
        };

        _mockApiClient.Setup(x => x.GetReplicationJobsAsync())
            .ReturnsAsync(expectedJobs);

        await _viewModel.LoadDataAsync();

        Assert.Equal(2, _viewModel.ReplicationJobs.Count);
        Assert.Contains(_viewModel.ReplicationJobs, j => j.Name == "VM Replication 1");
    }

    [Fact]
    public async Task LoadDataAsync_WhenError_SetsStatusMessage()
    {
        _mockApiClient.Setup(x => x.GetReplicationJobsAsync())
            .ThrowsAsync(new HttpRequestException("API Error"));

        await _viewModel.LoadDataAsync();

        Assert.Contains("Error", _viewModel.StatusMessage, StringComparison.OrdinalIgnoreCase);
    }

    [Fact]
    public void SelectedJob_PropertyChange_Notifies()
    {
        var job = new ReplicationJobModel { Id = "1", Name = "Test Job" };
        bool propertyChanged = false;
        _viewModel.PropertyChanged += (s, e) =>
        {
            if (e.PropertyName == nameof(ReplicationViewModel.SelectedJob))
                propertyChanged = true;
        };

        _viewModel.SelectedJob = job;

        Assert.True(propertyChanged);
        Assert.Equal(job, _viewModel.SelectedJob);
    }
}
