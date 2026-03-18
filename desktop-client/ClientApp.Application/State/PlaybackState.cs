using ClientApp.Application.Models;

namespace ClientApp.Application.State;

/// <summary>
/// Holds the most recent playback state for the client runtime.
/// </summary>
public sealed class PlaybackState
{
    /// <summary>
    /// Gets or sets the latest playback snapshot.
    /// </summary>
    public PlaybackStateModel Current { get; set; } = new(false, null, null);
}
