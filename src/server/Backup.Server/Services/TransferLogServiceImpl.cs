using Backup.Contracts;
using Grpc.Core;

namespace Backup.Server.Services;

public class FileTransferServiceImpl : FileTransferService.FileTransferServiceBase
{
    private readonly ILogger<FileTransferServiceImpl> _logger;
    private readonly Dictionary<string, TransferProgress> _transfers = new();

    public FileTransferServiceImpl(ILogger<FileTransferServiceImpl> logger)
    {
        _logger = logger;
    }

    public override Task<TransferResponse> InitTransfer(TransferInit request, ServerCallContext context)
    {
        _transfers[request.TransferId] = new TransferProgress
        {
            TransferId = request.TransferId,
            Status = TransferStatus.TransferStatusPending,
            TotalBytes = request.TotalSize
        };
        _logger.LogInformation("Initialized transfer {TransferId}", request.TransferId);
        return Task.FromResult(new TransferResponse { Success = true, TransferId = request.TransferId });
    }

    public override async Task<TransferResponse> UploadStream(IAsyncStreamReader<Chunk> requestStream, ServerCallContext context)
    {
        await foreach (var chunk in requestStream.ReadAllAsync())
        {
            if (_transfers.TryGetValue(chunk.TransferId, out var progress))
            {
                progress.BytesTransferred += chunk.Data.Length;
                progress.Status = TransferStatus.TransferStatusInProgress;
            }
        }
        return new TransferResponse { Success = true };
    }

    public override async Task DownloadStream(TransferInit request, IServerStreamWriter<Chunk> responseStream, ServerCallContext context)
    {
        _logger.LogInformation("Starting download stream for {TransferId}", request.TransferId);
        await Task.CompletedTask;
    }

    public override Task<TransferProgress> GetProgress(TransferRequest request, ServerCallContext context)
    {
        if (_transfers.TryGetValue(request.TransferId, out var progress))
            return Task.FromResult(progress);
        throw new RpcException(new Status(StatusCode.NotFound, "Transfer not found"));
    }

    public override Task<TransferResponse> PauseTransfer(TransferRequest request, ServerCallContext context)
    {
        if (_transfers.TryGetValue(request.TransferId, out var progress))
            progress.Status = TransferStatus.TransferStatusPaused;
        return Task.FromResult(new TransferResponse { Success = true });
    }

    public override Task<TransferResponse> ResumeTransfer(TransferRequest request, ServerCallContext context)
    {
        if (_transfers.TryGetValue(request.TransferId, out var progress))
            progress.Status = TransferStatus.TransferStatusInProgress;
        return Task.FromResult(new TransferResponse { Success = true });
    }

    public override Task<TransferResponse> CancelTransfer(TransferRequest request, ServerCallContext context)
    {
        if (_transfers.TryGetValue(request.TransferId, out var progress))
            progress.Status = TransferStatus.TransferStatusCancelled;
        return Task.FromResult(new TransferResponse { Success = true });
    }
}

public class LogServiceImpl : LogService.LogServiceBase
{
    private readonly ILogger<LogServiceImpl> _logger;

    public LogServiceImpl(ILogger<LogServiceImpl> logger)
    {
        _logger = logger;
    }

    public override async Task SendLog(IAsyncStreamReader<LogMessage> requestStream, ServerCallContext context)
    {
        await foreach (var log in requestStream.ReadAllAsync())
        {
            _logger.LogInformation("[{Level}] {Message}", log.Level, log.Message);
        }
    }

    public override Task<LogQueryResponse> QueryLogs(LogQueryRequest request, ServerCallContext context)
    {
        return Task.FromResult(new LogQueryResponse { TotalCount = 0 });
    }
}
