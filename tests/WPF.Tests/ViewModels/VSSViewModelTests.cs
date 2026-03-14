using System;
using System.Collections.Generic;
using System.Net.Http;
using System.Threading.Tasks;
using Moq;
using NovaBackup.GUI.Services;
using NovaBackup.GUI.ViewModels;
using Xunit;

namespace NovaBackup.WPF.Tests.ViewModels;

public class VSSViewModelTests
{
    private readonly Mock<IApiClient> _mockApiClient;
    private readonly VSSViewModel _viewModel;

    public VSSViewModelTests()
    {
        _mockApiClient = new Mock<IApiClient>();
        _viewModel = new VSSViewModel(_mockApiClient.Object);
    }

    [Fact]
    public void Constructor_InitializesWithEmptyCollection()
    {
        Assert.NotNull(_viewModel.Writers);
        Assert.Empty(_viewModel.Writers);
        Assert.False(_viewModel.IsLoading);
    }

    [Fact]
    public async Task LoadDataAsync_WhenSuccess_UpdatesWritersCollection()
    {
        var expectedWriters = new List<VSSWriterModel>
        {
            new VSSWriterModel { Name = "VSS Writer 1", State = "Stable", WriterType = "System" },
            new VSSWriterModel { Name = "VSS Writer 2", State = "Stable", WriterType = "Application" }
        };

        _mockApiClient.Setup(x => x.GetVSSWritersAsync())
            .ReturnsAsync(expectedWriters);

        await _viewModel.LoadDataCommand.ExecuteAsync(null);

        Assert.Equal(2, _viewModel.Writers.Count);
        Assert.Contains(_viewModel.Writers, w => w.Name == "VSS Writer 1");
        Assert.Contains(_viewModel.Writers, w => w.Name == "VSS Writer 2");
    }

    [Fact]
    public async Task LoadDataAsync_WhenError_SetsStatusMessage()
    {
        _mockApiClient.Setup(x => x.GetVSSWritersAsync())
            .ThrowsAsync(new HttpRequestException("API Error"));

        await _viewModel.LoadDataCommand.ExecuteAsync(null);

        Assert.Contains("Error", _viewModel.StatusMessage, StringComparison.OrdinalIgnoreCase);
    }

    [Fact]
    public async Task LoadDataAsync_SetsLoadingFlag()
    {
        var tcs = new TaskCompletionSource<List<VSSWriterModel>>();
        _mockApiClient.Setup(x => x.GetVSSWritersAsync())
            .Returns(tcs.Task);

        var task = _viewModel.LoadDataCommand.ExecuteAsync(null);

        Assert.True(_viewModel.IsLoading);

        tcs.SetResult(new List<VSSWriterModel>());
        await task;

        Assert.False(_viewModel.IsLoading);
    }
}
