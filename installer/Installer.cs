using System;
using System.Diagnostics;
using System.IO;
using System.IO.Compression;
using System.Reflection;
using System.Windows.Forms;

class Installer
{
    static void Main()
    {
        if (!IsAdmin())
        {
            MessageBox.Show("Administrator rights required!\n\nPlease right-click this file and select 'Run as administrator'",
                "NovaBackup Setup", MessageBoxButtons.OK, MessageBoxIcon.Error);
            return;
        }

        var result = MessageBox.Show("Install NovaBackup Enterprise v6.0?\n\nThis will:\n• Install to C:\\Program Files\\NovaBackup\n• Install Windows Service\n• Create desktop shortcut",
            "NovaBackup Setup", MessageBoxButtons.YesNo, MessageBoxIcon.Question);

        if (result != DialogResult.Yes) return;

        try
        {
            string installDir = @"C:\Program Files\NovaBackup";
            string dataDir = @"C:\ProgramData\NovaBackup";

            // Stop existing service
            RunCommand("net", "stop NovaBackup");
            RunCommand("sc", "delete NovaBackup");
            System.Threading.Thread.Sleep(2000);

            // Create directories
            Directory.CreateDirectory(installDir);
            Directory.CreateDirectory(Path.Combine(dataDir, "Logs"));
            Directory.CreateDirectory(Path.Combine(dataDir, "Backups"));
            Directory.CreateDirectory(Path.Combine(dataDir, "Config"));

            // Extract files from embedded ZIP
            string tempZip = Path.Combine(Path.GetTempPath(), "NovaBackup.zip");
            File.WriteAllBytes(tempZip, GetEmbeddedZip());
            ZipFile.ExtractToDirectory(tempZip, installDir);
            File.Delete(tempZip);

            // Install service
            RunCommand(Path.Combine(installDir, "nova.exe"), "install");
            RunCommand(Path.Combine(installDir, "nova.exe"), "start");

            // Create shortcuts
            CreateShortcut("NovaBackup Enterprise", Path.Combine(installDir, "NovaBackup.exe"),
                Environment.GetFolderPath(Environment.SpecialFolder.Desktop));
            CreateShortcut("NovaBackup Enterprise", Path.Combine(installDir, "NovaBackup.exe"),
                Environment.GetFolderPath(Environment.SpecialFolder.StartMenu));

            MessageBox.Show("Installation complete!\n\nNovaBackup Enterprise v6.0 has been installed.",
                "NovaBackup Setup", MessageBoxButtons.OK, MessageBoxIcon.Information);

            Process.Start(Path.Combine(installDir, "NovaBackup.exe"));
        }
        catch (Exception ex)
        {
            MessageBox.Show($"Installation failed:\n{ex.Message}", "Error",
                MessageBoxButtons.OK, MessageBoxIcon.Error);
        }
    }

    static bool IsAdmin()
    {
        try
        {
            using (var f = File.OpenWrite("\\\\.\\PHYSICALDRIVE0")) return true;
        }
        catch { return false; }
    }

    static void RunCommand(string file, string args)
    {
        var psi = new ProcessStartInfo(file, args)
        {
            UseShellExecute = false,
            CreateNoWindow = true,
            WindowStyle = ProcessWindowStyle.Hidden
        };
        Process.Start(psi)?.WaitForExit();
    }

    static void CreateShortcut(string name, string target, string folder)
    {
        try
        {
            var shell = new Microsoft.VisualBasic.Interaction();
            // Using WScript.Shell via reflection
            var wsh = Activator.CreateInstance(Type.GetTypeFromProgID("WScript.Shell"));
            var shortcut = wsh.GetType().InvokeMember("CreateShortcut",
                BindingFlags.InvokeMethod, null, wsh, new object[] {
                    Path.Combine(folder, name + ".lnk")
                });
            shortcut.GetType().InvokeMember("TargetPath",
                BindingFlags.SetProperty, null, shortcut, new object[] { target });
            shortcut.GetType().InvokeMember("WorkingDirectory",
                BindingFlags.SetProperty, null, shortcut, new object[] { Path.GetDirectoryName(target) });
            shortcut.GetType().InvokeMember("Save",
                BindingFlags.InvokeMethod, null, shortcut, null);
        }
        catch { /* Ignore shortcut creation errors */ }
    }

    static byte[] GetEmbeddedZip()
    {
        // This would contain the embedded ZIP data
        return new byte[0];
    }
}
