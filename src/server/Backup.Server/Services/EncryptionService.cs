using System.Security.Cryptography;
using Microsoft.Extensions.Configuration;

namespace Backup.Server.Services;

public interface IEncryptionService
{
    string Encrypt(string plainText);
    string Decrypt(string cipherText);
}

public class EncryptionService : IEncryptionService
{
    private readonly byte[] _key;
    private const int NonceSize = 12; // GCM recommended nonce size
    private const int TagSize = 16;   // GCM standard tag size

    public EncryptionService(IConfiguration configuration)
    {
        var envKey = Environment.GetEnvironmentVariable("ENCRYPTION_KEY");
        if (!string.IsNullOrEmpty(envKey))
        {
            _key = Convert.FromBase64String(envKey);
            return;
        }

        var keyFilePath = configuration["Encryption:KeyFilePath"];
        // ... (решта логіки)

        var keyFileDir = Path.Combine(AppContext.BaseDirectory, "data");
        Directory.CreateDirectory(keyFileDir);
        var defaultKeyPath = Path.Combine(keyFileDir, "encryption.key");

        if (File.Exists(defaultKeyPath))
        {
            var keyBytes = File.ReadAllBytes(defaultKeyPath);
            if (keyBytes.Length >= 32)
            {
                _key = keyBytes[..32];
                return;
            }
        }

        using var rng = RandomNumberGenerator.Create();
        var newKeyBytes = new byte[32];
        rng.GetBytes(newKeyBytes);
        File.WriteAllBytes(defaultKeyPath, newKeyBytes);
        File.SetAttributes(defaultKeyPath, FileAttributes.Hidden);

        _key = newKeyBytes;
    }

    public string Encrypt(string plainText)
    {
        if (string.IsNullOrEmpty(plainText))
            return plainText;

        var plainBytes = System.Text.Encoding.UTF8.GetBytes(plainText);
        var nonce = new byte[NonceSize];
        var tag = new byte[TagSize];
        var cipherBytes = new byte[plainBytes.Length];

        RandomNumberGenerator.Fill(nonce);

        using var aesGcm = new AesGcm(_key, TagSize);
        aesGcm.Encrypt(nonce, plainBytes, cipherBytes, tag);

        // Combined: Nonce (12) + Tag (16) + Ciphertext (N)
        var result = new byte[NonceSize + TagSize + cipherBytes.Length];
        Buffer.BlockCopy(nonce, 0, result, 0, NonceSize);
        Buffer.BlockCopy(tag, 0, result, NonceSize, TagSize);
        Buffer.BlockCopy(cipherBytes, 0, result, NonceSize + TagSize, cipherBytes.Length);

        return Convert.ToBase64String(result);
    }

    public string Decrypt(string cipherText)
    {
        if (string.IsNullOrEmpty(cipherText))
            return cipherText;

        var fullBytes = Convert.FromBase64String(cipherText);
        
        if (fullBytes.Length < NonceSize + TagSize)
            throw new CryptographicException("Invalid ciphertext format");

        var nonce = new byte[NonceSize];
        var tag = new byte[TagSize];
        var cipherBytes = new byte[fullBytes.Length - NonceSize - TagSize];

        Buffer.BlockCopy(fullBytes, 0, nonce, 0, NonceSize);
        Buffer.BlockCopy(fullBytes, NonceSize, tag, 0, TagSize);
        Buffer.BlockCopy(fullBytes, NonceSize + TagSize, cipherBytes, 0, cipherBytes.Length);

        var plainBytes = new byte[cipherBytes.Length];

        using var aesGcm = new AesGcm(_key, TagSize);
        aesGcm.Decrypt(nonce, cipherBytes, tag, plainBytes);

        return System.Text.Encoding.UTF8.GetString(plainBytes);
    }
}
