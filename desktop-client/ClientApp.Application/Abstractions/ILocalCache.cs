namespace ClientApp.Application.Abstractions;

/// <summary>
/// Stores and retrieves client cache data used for offline startup.
/// </summary>
public interface ILocalCache
{
    /// <summary>
    /// Loads the last cached schedule snapshot, if present.
    /// </summary>
    /// <param name="cancellationToken">Cancels the load request.</param>
    /// <returns>The cached payload or <see langword="null"/>.</returns>
    Task<string?> LoadScheduleSnapshotAsync(CancellationToken cancellationToken = default);

    /// <summary>
    /// Stores the latest serialized schedule snapshot.
    /// </summary>
    /// <param name="scheduleSnapshot">Serialized schedule state.</param>
    /// <param name="cancellationToken">Cancels the persistence operation.</param>
    Task SaveScheduleSnapshotAsync(string scheduleSnapshot, CancellationToken cancellationToken = default);
}
