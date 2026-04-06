using Backup.Server.Database;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Infrastructure;
using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace Backup.Server.Migrations;

[DbContext(typeof(BackupDbContext))]
[Migration("20260406000000_InitialCreate")]
partial class InitialCreate
{
}
