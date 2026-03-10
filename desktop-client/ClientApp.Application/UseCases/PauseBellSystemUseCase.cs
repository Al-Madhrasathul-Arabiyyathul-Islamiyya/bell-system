using ClientApp.Application.Abstractions;

namespace ClientApp.Application.UseCases;

/// <summary>
/// Stops current playback as part of a global pause action.
/// </summary>
public sealed class PauseBellSystemUseCase(IAudioPlaybackCoordinator audioPlaybackCoordinator)
{
    /// <summary>
    /// Executes the pause operation.
    /// </summary>
    /// <param name="cancellationToken">Cancels the operation.</param>
    public Task ExecuteAsync(CancellationToken cancellationToken = default) =>
        audioPlaybackCoordinator.StopAsync(cancellationToken);
}
