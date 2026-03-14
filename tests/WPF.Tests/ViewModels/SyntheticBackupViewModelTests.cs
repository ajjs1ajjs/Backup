using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using Moq;
using NovaBackup.GUI.Services;
using NovaBackup.GUI.ViewModels;
using Xunit;

namespace NovaBackup.WPF.Tests.ViewModels;

public class SyntheticBackupViewModelTests
{
    private readonly Mock<IApiClient> _mockApiClient;
    private readonly SyntheticBackupViewModel _viewModel;

    public SyntheticBackupViewModelTests()
    {
        _mockApiClient = new Mock<IApiClient>();
        _viewModel = new SyntheticBackupViewModel(_mockApiClient.Object);
    }

    [Fact]
    public void Constructor_InitializesWithEmptyCollection()
    {
        Assert.NotNull(_viewModel.Backups);
        Assert.Empty(_viewModel.Backups);
        Assert.False(_viewModel.IsLoading);
    }

    [Fact]
    public async Task LoadDataAsync_WhenSuccess_UpdatesBackupsCollection()
    {
        var expectedBackups = new List<SyntheticBackupModel>
        {
            new SyntheticBackupModel { Id = "1", Name = "Weekly Full", SourceRepo = "Repo-01", Status = "Completed" },
            new SyntheticBackupModel { Id = "2", Name = "Monthly Full", SourceRepo = "Repo-02", Status = "Running" }
        };

        _mockApiClient.Setup(x => x.GetSyntheticBackupsAsync())
            .ReturnsAsync(expectedBackups);

        await _viewModel.LoadDataCommand.ExecuteAsync(null);

        Assert.Equal(2, _viewModel.Backups.Count);
        Assert.Contains(_viewModel.Backups, b => b.Name == "Weekly Full");
    }

    [Fact]
    public async Task CreateSyntheticBackupAsync_WhenSuccess_ReloadsData()
    {
        _mockApiClient.Setup(x => x.CreateSyntheticBackupAsync(It.IsAny<SyntheticBackupRequest>()))
            .ReturnsAsync(true);
        _mockApiClient.Setup(x => x.GetSyntheticBackupsAsync())
            .ReturnsAsync(new List<SyntheticBackupModel> { new SyntheticBackupModel { Id = "1", Name = "New Backup" } });

        await _viewModel.CreateSyntheticBackupCommand.ExecuteAsync(null);

        _mockApiClient.Verify(x => x.CreateSyntheticBackupAsync(It.IsAny<SyntheticBackupRequest>()), Times.Once);
    }

    [Fact]
    public async Task DeleteSyntheticBackupAsync_WhenSuccess_RemovesFromCollection()
    {
        var backup = new SyntheticBackupModel { Id = "1", Name = "Test Backup" };
        _viewModel.Backups.Add(backup);
        _viewModel.SelectedBackup = backup;

        _mockApiClient.Setup(x => x.DeleteSyntheticBackupAsync("1"))
            .ReturnsAsync(true);
        _mockApiClient.Setup(x => x.GetSyntheticBackupsAsync())
            .ReturnsAsync(new List<SyntheticBackupModel>());

        await _viewModel.DeleteSyntheticBackupCommand.ExecuteAsync(null);

        Assert.Empty(_viewModel.Backups);
    }

    [Fact]
    public async Task MergeIncrementalsAsync_WhenSuccess_ReloadsData()
    {
        var backup = new SyntheticBackupModel { Id = "1", Name = "Test Backup" };
        _viewModel.SelectedBackup = backup;

        _mockApiClient.Setup(x => x.MergeIncrementalsAsync(It.IsAny<MergeIncrementalsRequest>()))
            .ReturnsAsync(true);
        _mockApiClient.Setup(x => x.GetSyntheticBackupsAsync())
            .ReturnsAsync(new List<SyntheticBackupModel> { backup });

        await _viewModel.MergeIncrementalsCommand.ExecuteAsync(null);

        _mockApiClient.Verify(x => x.MergeIncrementalsAsync(It.IsAny<MergeIncrementalsRequest>()), Times.Once);
    }

    [Fact]
    public async Task LoadDataAsync_WhenError_SetsStatusMessage()
    {
        _mockApiClient.Setup(x => x.GetSyntheticBackupsAsync())
            .ThrowsAsync(new Exception("API Error"));

        await _viewModel.LoadDataCommand.ExecuteAsync(null);

        Assert.Contains("Error", _viewModel.StatusMessage, StringComparison.OrdinalIgnoreCase);
    }

    [Fact]
    public void SelectedBackup_PropertyChange_Notifies()
    {
        var backup = new SyntheticBackupModel { Id = "1", Name = "Test" };
        bool propertyChanged = false;
        _viewModel.PropertyChanged += (s, e) =>
        {
            if (e.PropertyName == nameof(SyntheticBackupViewModel.SelectedBackup))
                propertyChanged = true;
        };

        _viewModel.SelectedBackup = backup;

        Assert.True(propertyChanged);
        Assert.Equal(backup, _viewModel.SelectedBackup);
    }
}
