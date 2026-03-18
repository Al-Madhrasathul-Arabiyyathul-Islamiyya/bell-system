namespace ClientApp.Application.Models;

/// <summary>
/// Describes the current client playback status.
/// </summary>
/// <param name="IsPlaying">Whether audio is currently playing.</param>
/// <param name="CurrentSoundName">The display name of the active sound.</param>
/// <param name="StartedAt">When the current playback started.</param>
public sealed record PlaybackStateModel(bool IsPlaying, string? CurrentSoundName, DateTimeOffset? StartedAt);
