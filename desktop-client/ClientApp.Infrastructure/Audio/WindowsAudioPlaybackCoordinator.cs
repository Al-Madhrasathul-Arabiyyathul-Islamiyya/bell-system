using ClientApp.Application.Abstractions;

namespace ClientApp.Infrastructure.Audio;

/// <summary>
/// Placeholder Windows audio playback implementation.
/// </summary>
public sealed class WindowsAudioPlaybackCoordinator : IAudioPlaybackCoordinator
{
    /// <inheritdoc />
    public Task PlayAsync(Guid mediaId, CancellationToken cancellationToken = default)
    {
        _ = mediaId;
        _ = cancellationToken;
        return Task.CompletedTask;
    }

    /// <inheritdoc />
    public Task StopAsync(CancellationToken cancellationToken = default)
    {
        _ = cancellationToken;
        return Task.CompletedTask;
    }
}
