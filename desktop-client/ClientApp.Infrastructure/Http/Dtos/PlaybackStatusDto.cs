namespace ClientApp.Infrastructure.Http.Dtos;

/// <summary>
/// Represents playback status returned by the backend API.
/// </summary>
public sealed class PlaybackStatusDto
{
    /// <summary>
    /// Gets or sets whether playback is active.
    /// </summary>
    public bool IsPlaying { get; set; }

    /// <summary>
    /// Gets or sets the current sound name.
    /// </summary>
    public string? CurrentSoundName { get; set; }
}
