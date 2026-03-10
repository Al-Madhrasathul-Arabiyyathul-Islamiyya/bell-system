using ClientApp.Application.Abstractions;

namespace ClientApp.Infrastructure.Persistence;

/// <summary>
/// Placeholder SQLite-backed cache implementation.
/// </summary>
public sealed class SqliteLocalCache : ILocalCache, ILocalScheduleCache
{
    /// <inheritdoc />
    public Task<string?> LoadScheduleSnapshotAsync(CancellationToken cancellationToken = default)
    {
        _ = cancellationToken;
        return Task.FromResult<string?>(null);
    }

    /// <inheritdoc />
    public Task SaveAsync(string scheduleSnapshot, CancellationToken cancellationToken = default)
    {
        _ = scheduleSnapshot;
        _ = cancellationToken;
        return Task.CompletedTask;
    }

    /// <inheritdoc />
    public Task SaveScheduleSnapshotAsync(string scheduleSnapshot, CancellationToken cancellationToken = default) =>
        SaveAsync(scheduleSnapshot, cancellationToken);
}
