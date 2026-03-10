using ClientApp.Application.Abstractions;

namespace ClientApp.Infrastructure.Services;

/// <summary>
/// Placeholder HTTP-backed schedule API implementation.
/// </summary>
public sealed class HttpScheduleApi : IScheduleApi
{
    /// <inheritdoc />
    public Task<string> GetScheduleSummaryAsync(CancellationToken cancellationToken = default) =>
        Task.FromResult("No schedule loaded.");
}
