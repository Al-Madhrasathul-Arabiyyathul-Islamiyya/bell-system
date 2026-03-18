namespace ClientApp.Application.Abstractions;

/// <summary>
/// Controls bell and media playback for the client runtime.
/// </summary>
public interface IAudioPlaybackCoordinator
{
    /// <summary>
    /// Starts playback for the requested media identifier.
    /// </summary>
    /// <param name="mediaId">The media item to play.</param>
    /// <param name="cancellationToken">Cancels the playback request.</param>
    Task PlayAsync(Guid mediaId, CancellationToken cancellationToken = default);

    /// <summary>
    /// Stops the active playback session, if any.
    /// </summary>
    /// <param name="cancellationToken">Cancels the stop request.</param>
    Task StopAsync(CancellationToken cancellationToken = default);
}
