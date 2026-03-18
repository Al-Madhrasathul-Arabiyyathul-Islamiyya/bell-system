namespace ClientApp.Application.Abstractions;

/// <summary>
/// Starts the realtime update channel used to receive schedule and playback changes.
/// </summary>
public interface IRealtimeUpdates
{
    /// <summary>
    /// Connects to the realtime update source.
    /// </summary>
    /// <param name="cancellationToken">Cancels the startup operation.</param>
    Task StartAsync(CancellationToken cancellationToken = default);
}
