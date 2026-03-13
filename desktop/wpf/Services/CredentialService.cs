using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using NovaBackup.GUI.Models;

namespace NovaBackup.GUI.Services
{
    /// <summary>
    /// Service for managing credentials with caching and validation
    /// </summary>
    public class CredentialService
    {
        private readonly IApiClient _apiClient;
        private List<CredentialModel> _cachedCredentials = new();
        private DateTime _cacheTimestamp;
        private readonly TimeSpan _cacheDuration = TimeSpan.FromMinutes(5);

        public CredentialService(IApiClient apiClient)
        {
            _apiClient = apiClient;
        }

        /// <summary>
        /// Get all credentials with caching
        /// </summary>
        public async Task<List<CredentialModel>> GetCredentialsAsync(bool forceRefresh = false)
        {
            if (!forceRefresh && _cachedCredentials.Any() && DateTime.Now - _cacheTimestamp < _cacheDuration)
            {
                return _cachedCredentials;
            }

            try
            {
                var credentials = await _apiClient.GetCredentialsAsync();
                _cachedCredentials = credentials;
                _cacheTimestamp = DateTime.Now;
                return credentials;
            }
            catch (Exception)
            {
                // Return cached data on error if available
                if (_cachedCredentials.Any())
                {
                    return _cachedCredentials;
                }
                throw;
            }
        }

        /// <summary>
        /// Get credential by ID
        /// </summary>
        public CredentialModel? GetCredentialById(string id)
        {
            return _cachedCredentials.FirstOrDefault(c => c.Id == id);
        }

        /// <summary>
        /// Get credentials by type
        /// </summary>
        public List<CredentialModel> GetCredentialsByType(string type)
        {
            return _cachedCredentials.Where(c => c.Type == type).ToList();
        }

        /// <summary>
        /// Create new credential
        /// </summary>
        public async Task<bool> CreateCredentialAsync(CredentialModel credential)
        {
            var success = await _apiClient.CreateCredentialAsync(credential);
            if (success)
            {
                await GetCredentialsAsync(forceRefresh: true);
            }
            return success;
        }

        /// <summary>
        /// Delete credential
        /// </summary>
        public async Task<bool> DeleteCredentialAsync(string id)
        {
            var success = await _apiClient.DeleteCredentialAsync(id);
            if (success)
            {
                await GetCredentialsAsync(forceRefresh: true);
            }
            return success;
        }

        /// <summary>
        /// Validate credential format
        /// </summary>
        public static bool ValidateCredential(CredentialModel credential, out string errorMessage)
        {
            if (string.IsNullOrWhiteSpace(credential.Name))
            {
                errorMessage = "Name is required";
                return false;
            }

            if (string.IsNullOrWhiteSpace(credential.Username))
            {
                errorMessage = "Username is required";
                return false;
            }

            if (string.IsNullOrWhiteSpace(credential.Type))
            {
                errorMessage = "Type is required";
                return false;
            }

            errorMessage = string.Empty;
            return true;
        }

        /// <summary>
        /// Clear cache
        /// </summary>
        public void ClearCache()
        {
            _cachedCredentials.Clear();
            _cacheTimestamp = DateTime.MinValue;
        }
    }
}
