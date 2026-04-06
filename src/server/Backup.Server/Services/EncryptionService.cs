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
    private readonly byte[] _iv;

    public EncryptionService(IConfiguration configuration)
    {
        var keyFilePath = configuration["Encryption:KeyFilePath"];

        if (!string.IsNullOrWhiteSpace(keyFilePath) && File.Exists(keyFilePath))
        {
            var keyBytes = File.ReadAllBytes(keyFilePath);
            if (keyBytes.Length >= 48)
            {
                _key = keyBytes[..32];
                _iv = keyBytes[32..48];
                return;
            }
        }

        var keyFileDir = Path.Combine(AppContext.BaseDirectory, "data");
        Directory.CreateDirectory(keyFileDir);
        var defaultKeyPath = Path.Combine(keyFileDir, "encryption.key");

        if (File.Exists(defaultKeyPath))
        {
            var keyBytes = File.ReadAllBytes(defaultKeyPath);
            if (keyBytes.Length >= 48)
            {
                _key = keyBytes[..32];
                _iv = keyBytes[32..48];
                return;
            }
        }

        using var rng = RandomNumberGenerator.Create();
        var newKeyBytes = new byte[48];
        rng.GetBytes(newKeyBytes);
        File.WriteAllBytes(defaultKeyPath, newKeyBytes);
        File.SetAttributes(defaultKeyPath, FileAttributes.Hidden);

        _key = newKeyBytes[..32];
        _iv = newKeyBytes[32..48];
    }

    public string Encrypt(string plainText)
    {
        if (string.IsNullOrEmpty(plainText))
            return plainText;

        using var aes = Aes.Create();
        aes.Key = _key;
        aes.IV = _iv;
        aes.Mode = CipherMode.CBC;
        aes.Padding = PaddingMode.PKCS7;

        using var encryptor = aes.CreateEncryptor(aes.Key, aes.IV);
        var plainBytes = System.Text.Encoding.UTF8.GetBytes(plainText);
        var cipherBytes = encryptor.TransformFinalBlock(plainBytes, 0, plainBytes.Length);

        return Convert.ToBase64String(cipherBytes);
    }

    public string Decrypt(string cipherText)
    {
        if (string.IsNullOrEmpty(cipherText))
            return cipherText;

        using var aes = Aes.Create();
        aes.Key = _key;
        aes.IV = _iv;
        aes.Mode = CipherMode.CBC;
        aes.Padding = PaddingMode.PKCS7;

        using var decryptor = aes.CreateDecryptor(aes.Key, aes.IV);
        var cipherBytes = Convert.FromBase64String(cipherText);
        var plainBytes = decryptor.TransformFinalBlock(cipherBytes, 0, cipherBytes.Length);

        return System.Text.Encoding.UTF8.GetString(plainBytes);
    }
}
