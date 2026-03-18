using ClientApp.Application.Abstractions;

namespace ClientApp.Application.UseCases;

/// <summary>
/// Starts a manually requested playback item.
/// </summary>
public sealed class TriggerManualPlaybackUseCase(IAudioPlaybackCoordinator audioPlaybackCoordinator)
{
    /// <summary>
    /// Plays the requested media item.
    /// </summary>
    /// <param name="mediaId">The media identifier to play.</param>
    /// <param name="cancellationToken">Cancels the operation.</param>
    public Task ExecuteAsync(Guid mediaId, CancellationToken cancellationToken = default) =>
        audioPlaybackCoordinator.PlayAsync(mediaId, cancellationToken);
}
