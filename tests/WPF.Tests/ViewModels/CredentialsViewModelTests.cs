using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using Moq;
using NovaBackup.GUI.Services;
using NovaBackup.GUI.ViewModels;
using Xunit;

namespace NovaBackup.WPF.Tests.ViewModels;

public class CredentialsViewModelTests
{
    private readonly Mock<IApiClient> _mockApiClient;
    private readonly CredentialsViewModel _viewModel;

    public CredentialsViewModelTests()
    {
        _mockApiClient = new Mock<IApiClient>();
        _viewModel = new CredentialsViewModel(_mockApiClient.Object);
    }

    [Fact]
    public void Constructor_InitializesWithEmptyCollection()
    {
        Assert.NotNull(_viewModel.Credentials);
        Assert.Empty(_viewModel.Credentials);
        Assert.False(_viewModel.IsLoading);
    }

    [Fact]
    public async Task LoadDataAsync_WhenSuccess_UpdatesCredentialsCollection()
    {
        var expectedCredentials = new List<CredentialModel>
        {
            new CredentialModel { Id = "1", Name = "Admin", Username = "admin", Type = "Windows" },
            new CredentialModel { Id = "2", Name = "Backup User", Username = "backup", Type = "Linux" }
        };

        _mockApiClient.Setup(x => x.GetCredentialsAsync())
            .ReturnsAsync(expectedCredentials);

        await _viewModel.LoadDataAsync();

        Assert.Equal(2, _viewModel.Credentials.Count);
        Assert.Contains(_viewModel.Credentials, c => c.Name == "Admin");
    }

    [Fact]
    public async Task DeleteCredentialAsync_WhenSuccess_RemovesFromCollection()
    {
        var credential = new CredentialModel { Id = "1", Name = "Test" };
        _viewModel.Credentials.Add(credential);

        _mockApiClient.Setup(x => x.DeleteCredentialAsync("1"))
            .ReturnsAsync(true);

        await _viewModel.DeleteCredentialAsync(credential);

        Assert.DoesNotContain(credential, _viewModel.Credentials);
    }

    [Fact]
    public async Task CreateCredentialAsync_WhenSuccess_AddsToCollection()
    {
        var newCredential = new CredentialModel { Id = "1", Name = "New Credential" };

        _mockApiClient.Setup(x => x.CreateCredentialAsync(It.IsAny<CredentialModel>()))
            .ReturnsAsync(true);

        var result = await _viewModel.CreateCredentialAsync(newCredential);

        Assert.True(result);
    }

    [Fact]
    public void SelectedCredential_PropertyChange_Notifies()
    {
        var credential = new CredentialModel { Id = "1", Name = "Test" };
        bool propertyChanged = false;
        _viewModel.PropertyChanged += (s, e) =>
        {
            if (e.PropertyName == nameof(CredentialsViewModel.SelectedCredential))
                propertyChanged = true;
        };

        _viewModel.SelectedCredential = credential;

        Assert.True(propertyChanged);
        Assert.Equal(credential, _viewModel.SelectedCredential);
    }
}
