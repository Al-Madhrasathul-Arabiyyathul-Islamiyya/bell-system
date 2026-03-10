namespace ClientApp.Application.Abstractions;

/// <summary>
/// Persists a local copy of the current schedule state for offline startup.
/// </summary>
public interface ILocalScheduleCache
{
    /// <summary>
    /// Stores the latest serialized schedule snapshot.
    /// </summary>
    /// <param name="scheduleSnapshot">Serialized schedule state.</param>
    /// <param name="cancellationToken">Cancels the persistence operation.</param>
    Task SaveAsync(string scheduleSnapshot, CancellationToken cancellationToken = default);
}
