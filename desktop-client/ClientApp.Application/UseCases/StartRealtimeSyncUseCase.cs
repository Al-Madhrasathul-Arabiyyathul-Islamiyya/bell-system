using ClientApp.Application.Abstractions;

namespace ClientApp.Application.UseCases;

/// <summary>
/// Starts the realtime synchronization channel for schedule updates.
/// </summary>
public sealed class StartRealtimeSyncUseCase(IRealtimeUpdates realtimeUpdates)
{
    /// <summary>
    /// Opens the realtime connection.
    /// </summary>
    /// <param name="cancellationToken">Cancels the operation.</param>
    public Task ExecuteAsync(CancellationToken cancellationToken = default) =>
        realtimeUpdates.StartAsync(cancellationToken);
}
