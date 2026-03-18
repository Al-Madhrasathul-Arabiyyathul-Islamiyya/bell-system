using ClientApp.Application.Abstractions;

namespace ClientApp.Infrastructure.Services;

/// <summary>
/// Placeholder local cache implementation for persisted schedule data.
/// </summary>
public sealed class LocalScheduleCache : ILocalScheduleCache
{
    /// <inheritdoc />
    public Task SaveAsync(string scheduleSnapshot, CancellationToken cancellationToken = default) =>
        Task.CompletedTask;
}
